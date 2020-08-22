package controller

import (
	"encoding/json"
	"errors"
	"github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/postoffice"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/apple"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/pkg/letter"
	"github.com/FTChinese/subscription-api/repository/iaprepo"
	"github.com/FTChinese/subscription-api/repository/readerrepo"
	"github.com/jmoiron/sqlx"
	"io/ioutil"
	"net/http"
)

type IAPRouter struct {
	secret     string
	config     config.BuildConfig
	iapRepo    iaprepo.Env
	readerRepo readerrepo.ReaderEnv
	postman    postoffice.PostOffice
}

func NewIAPRouter(db *sqlx.DB, cfg config.BuildConfig, p postoffice.PostOffice) IAPRouter {
	return IAPRouter{
		secret:     config.MustIAPSecret(),
		config:     cfg,
		iapRepo:    iaprepo.NewEnv(db, cfg),
		readerRepo: readerrepo.NewReaderEnv(db, cfg),
		postman:    p,
	}
}

// Verify a receipt and get response from app store.
// This the shared action when verifying receipt,
// link account and unlink account.
// See https://developer.apple.com/documentation/storekit/in-app_purchase/validating_receipts_with_the_app_store
func (router IAPRouter) doVerification(req *http.Request) (apple.VerificationResp, *render.ResponseError) {
	log := logger.WithField("trace", "IAPRouter.VerifyReceipt")

	// Parse input data.
	var payload apple.VerificationPayload
	if err := gorest.ParseJSON(req.Body, &payload); err != nil {
		log.Error(err)
		return apple.VerificationResp{}, render.NewBadRequest(err.Error())
	}

	// Validate input.
	if ve := payload.Validate(); ve != nil {
		return apple.VerificationResp{}, render.NewUnprocessable(ve)
	}
	// Set other requires fields of payload.
	payload.Password = router.secret
	payload.ExcludeOldTransactions = false

	// Verify
	resp, err := apple.VerifyReceipt(payload, router.config.IAPVerificationURL())

	if err != nil {
		return apple.VerificationResp{}, render.NewBadRequest(err.Error())
	}

	log.Infof("Environment %s, is retryable %t, status %d", resp.Environment, resp.IsRetryable, resp.Status)

	// If the response is not valid.
	if !resp.Validate() {
		log.Info("Verification response is not valid")

		log.WithField("environment", resp.Environment).
			WithField("status", resp.Status).
			WithField("message", resp.GetStatusMessage()).
			WithField("receiptData", payload.ReceiptData).
			Info("IAP verification failed")

		ve := &render.ValidationError{
			Message: "verification response is not valid",
			Field:   "receipt-data",
			Code:    render.CodeInvalid,
		}
		return apple.VerificationResp{}, render.NewUnprocessable(ve)
	}

	// Find the latest valid transaction.
	resp.Parse()

	// Save the decoded receipt.
	go func() {
		_ = router.iapRepo.SaveVerifiedReceipt(
			resp.ReceiptSchema(),
		)
	}()

	// Dissect and save other fields in the verification response.
	router.saveReceiptData(resp.UnifiedReceipt)
	return resp, nil
}

// Save various fields in apple.UnifiedReceipt:
// * LatestTransactions
// * PendingRenewalInfo
// * An encoded receipt as a file.
func (router IAPRouter) saveReceiptData(ur apple.UnifiedReceipt) {

	// Save the LatestTransactions array.
	go func() {
		for _, v := range ur.LatestTransactions {
			_ = router.iapRepo.SaveTransaction(
				v.Schema(ur.Environment),
			)
		}
	}()

	// Save PendingRenewalInfo array
	go func() {
		for _, v := range ur.PendingRenewalInfo {
			_ = router.iapRepo.SavePendingRenewal(
				v.Schema(ur.Environment),
			)
		}
	}()

	// Save the LatestToken field to a file.
	// Initially it was designed to save in SQL.
	receiptToken := ur.ReceiptToken()

	go func() {
		_ = iaprepo.SaveReceiptTokenFile(receiptToken)
	}()
}

func (router IAPRouter) createSubscription(resp apple.VerificationResp) (apple.Subscription, *render.ResponseError) {
	sub, err := resp.Subscription()
	if err != nil {
		return apple.Subscription{}, render.NewNotFound(err.Error())
	}

	err = router.iapRepo.CreateSubscription(sub)
	if err != nil {
		return apple.Subscription{}, render.NewDBError(err)
	}

	return sub, nil
}

// VerifyReceipt verifies if the receipt data send by client is valid.
// Input
// receipt-data: string
func (router IAPRouter) VerifyReceipt(w http.ResponseWriter, req *http.Request) {
	resp, err := router.doVerification(req)
	if err != nil {
		_ = render.New(w).JSON(err.StatusCode, err)
		return
	}

	_, err = router.createSubscription(resp)
	if err != nil {
		_ = render.New(w).JSON(err.StatusCode, err)
		return
	}

	_ = render.New(w).OK(resp)
}

// Link merges IAP subscription to FTC account.
func (router IAPRouter) Link(w http.ResponseWriter, req *http.Request) {
	// Get user's ids.
	memberID := getReaderIDs(req.Header)

	// Verification
	resp, resErr := router.doVerification(req)
	if resErr != nil {
		_ = render.New(w).JSON(resErr.StatusCode, resErr)
		return
	}

	// Build subscription.
	sub, resErr := router.createSubscription(resp)
	if resErr != nil {
		_ = render.New(w).JSON(resErr.StatusCode, resErr)
		return
	}

	// Start to link apple subscription to ftc membership.
	linkResult, err := router.iapRepo.Link(sub, memberID)

	if err != nil {
		var ve *render.ValidationError
		if errors.As(err, &ve) {
			_ = render.New(w).Unprocessable(ve)
			return
		}

		_ = render.New(w).DBError(err)
		return
	}

	if linkResult.IsInitialLink() {
		go func() {
			account, err := router.readerRepo.FindAccountByFtcID(linkResult.Linked.FtcID.String)
			if err != nil {
				return
			}

			parcel, err := letter.NewIAPLinkParcel(account, linkResult.Linked)
			if err != nil {
				return
			}

			err = router.postman.Deliver(parcel)
			if err != nil {
				return
			}
		}()
	}

	_ = render.New(w).OK(linkResult.Linked)
}

// Unlink removes apple subscription id from a user's membership
func (router IAPRouter) Unlink(w http.ResponseWriter, req *http.Request) {
	// Get user's ids.
	readerIDs := getReaderIDs(req.Header)

	// Verification
	resp, resErr := router.doVerification(req)
	if resErr != nil {
		_ = render.New(w).JSON(resErr.StatusCode, resErr)
		return
	}

	// Build subscription.
	sub, resErr := router.createSubscription(resp)
	if resErr != nil {
		_ = render.New(w).JSON(resErr.StatusCode, resErr)
		return
	}

	if err := router.iapRepo.Unlink(sub, readerIDs); err != nil {
		if errors.Is(err, apple.ErrUnlinkMismatchedFTC) {
			_ = render.New(w).BadRequest(err.Error())
			return
		}

		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).NoContent()
}

// WebHook receives app store server-to-server notification.
func (router IAPRouter) WebHook(w http.ResponseWriter, req *http.Request) {
	log := logger.WithField("trace", "IAPRouter.WebHook")

	var wh apple.WebHook
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Error(err)
		return
	}

	log.Infof("Apple webhook raw: %s", b)

	if err := json.Unmarshal(b, &wh); err != nil {
		log.Error(err)
		_ = render.New(w).BadRequest("")
		return
	}

	_ = router.iapRepo.SaveNotification(wh.Schema())

	if !wh.UnifiedReceipt.Validate() {
		log.Infof("invalid webhook unified receipt")
		_ = render.New(w).Unprocessable(&render.ValidationError{
			Message: "unified receipt field is not valid",
			Field:   "unified_receipt",
			Code:    render.CodeInvalid,
		})
		return
	}
	// Find the latest transaction and save transaction
	// history.
	wh.UnifiedReceipt.Parse()

	router.saveReceiptData(wh.UnifiedReceipt)

	// Build apple's subscription and save it.
	sub, err := wh.UnifiedReceipt.Subscription()
	if err != nil {
		log.Error(err)
		_ = render.New(w).BadRequest("")
		return
	}

	if err := router.iapRepo.CreateSubscription(sub); err != nil {
		log.Error(err)
		_ = render.New(w).BadRequest("")
		return
	}

	snapshot, err := router.iapRepo.WebhookUpdate(sub)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	// Take a snapshot.
	go func() {
		_ = router.iapRepo.BackUpMember(snapshot)
	}()

	_ = render.New(w).OK(nil)
}

// LoadReceipt retrieves the subscription data for
// an original transaction id, together with the
// receipt used to verify it.
func (router IAPRouter) LoadReceipt(w http.ResponseWriter, req *http.Request) {
	id, _ := getURLParam(req, "id").ToString()

	sub, err := router.iapRepo.LoadSubscription(id)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	b, err := iaprepo.LoadReceipt(sub.OriginalTransactionID, sub.Environment)
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	data := struct {
		apple.Subscription
		Receipt string `json:"receipt"`
	}{
		Subscription: sub,
		Receipt:      string(b),
	}

	_ = render.New(w).OK(data)
}
