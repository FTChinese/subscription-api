package controller

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/postoffice"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/apple"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/pkg/letter"
	"github.com/FTChinese/subscription-api/repository/iaprepo"
	"github.com/FTChinese/subscription-api/repository/readerrepo"
	"github.com/jmoiron/sqlx"
)

type IAPRouter struct {
	secret     string
	config     config.BuildConfig
	iapRepo    iaprepo.Env
	readerRepo readerrepo.Env
	postman    postoffice.PostOffice
	iapClient  iaprepo.Client
}

func NewIAPRouter(db *sqlx.DB, cfg config.BuildConfig, p postoffice.PostOffice) IAPRouter {
	return IAPRouter{
		secret:     config.MustIAPSecret(),
		config:     cfg,
		iapRepo:    iaprepo.NewEnv(db, cfg),
		readerRepo: readerrepo.NewEnv(db, cfg),
		postman:    p,
		iapClient:  iaprepo.NewClient(cfg.Sandbox()),
	}
}

// Verify a receipt and get response from app store.
// This the shared action when verifying receipt,
// link account and unlink account.
// See https://developer.apple.com/documentation/storekit/in-app_purchase/validating_receipts_with_the_app_store
func (router IAPRouter) doVerification(receipt string) (apple.VerificationResp, *render.ResponseError) {
	sugar := logger.Sugar()

	content, err := router.iapClient.Verify(receipt)
	if err != nil {
		return apple.VerificationResp{}, render.NewInternalError(err.Error())
	}

	sugar.Infof("Verified receipt %s", content)

	var resp apple.VerificationResp
	if err := json.Unmarshal(content, &resp); err != nil {
		return resp, render.NewInternalError(err.Error())
	}

	sugar.Infof("Environment %s, is retryable %t, status %d", resp.Environment, resp.IsRetryable, resp.Status)

	// If the response is not valid.
	if !resp.Validate() {
		sugar.Info("IAP verification response is not valid")

		ve := &render.ValidationError{
			Message: "verification response is not valid",
			Field:   "receiptData",
			Code:    render.CodeInvalid,
		}
		return apple.VerificationResp{}, render.NewUnprocessable(ve)
	}

	// Find the latest valid transaction.
	resp.Parse()

	// Save the decoded receipt.
	go func() {
		_ = router.iapRepo.SaveDecodedReceipt(
			resp.ReceiptSchema(),
		)
	}()

	// Dissect and save other fields in the verification response.
	router.iapRepo.SaveResponsePayload(resp.UnifiedReceipt)
	return resp, nil
}

// VerifyReceipt verifies if the receipt data send by client is valid. After app store responded,
// its latest_receipt, latest_receipt_info, pending_renewal_info are saved in DB in background thread.
// An apple.Subscription is created from the response, which is saved or updated if already exists,
// and then user's membership is updated if it exists.
//
// Input
// receipt-data: string
func (router IAPRouter) VerifyReceipt(w http.ResponseWriter, req *http.Request) {
	var input apple.ReceiptInput
	if err := gorest.ParseJSON(req.Body, &input); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}
	if ve := input.Validate(); ve != nil {
		_ = render.New(w).Unprocessable(ve)
		return
	}

	resp, resErr := router.doVerification(input.ReceiptData)
	if resErr != nil {
		_ = render.New(w).JSON(resErr.StatusCode, resErr)
		return
	}

	sub, err := resp.Subscription()

	// Update subscription and possible membership in background since this step is irrelevant to verification.
	if err == nil {
		go func() {
			_ = router.iapRepo.UpsertSubscription(sub)

			snapshot, err := router.iapRepo.UpdateMembership(sub)
			if err != nil {
				return
			}

			if !snapshot.IsZero() {
				_ = router.readerRepo.BackUpMember(snapshot)
				return
			}
		}()
	}

	_ = render.New(w).OK(resp)
}

// Link merges IAP subscription to FTC account.
//
// Input:
// ftcId: string;
// originalTxId: string;
func (router IAPRouter) Link(w http.ResponseWriter, req *http.Request) {
	var input apple.LinkInput
	if err := gorest.ParseJSON(req.Body, &input); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}
	if ve := input.Validate(); ve != nil {
		_ = render.New(w).Unprocessable(ve)
		return
	}

	ftcAccount, err := router.readerRepo.FtcAccountByFtcID(input.FtcID)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	iapSubs, err := router.iapRepo.LoadSubscription(input.OriginalTxID)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	// Start to link apple subscription to ftc membership.
	linkResult, err := router.iapRepo.Link(ftcAccount, iapSubs)

	if err != nil {
		var ve *render.ValidationError
		// ValidationError indicates the link is not allowed.
		if errors.As(err, &ve) {
			_ = render.New(w).Unprocessable(ve)
			return
		}

		_ = render.New(w).DBError(err)
		return
	}

	// Send notification email if this is initial link.
	if linkResult.Initial {
		go func() {
			account, err := router.readerRepo.FtcAccountByFtcID(linkResult.Linked.FtcID.String)
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
//
// Input:
// originalTxId: string
// ftcId: string
func (router IAPRouter) Unlink(w http.ResponseWriter, req *http.Request) {
	// Get user's ids.
	readerIDs := getReaderIDs(req.Header)

	var s apple.BaseSchema
	if err := gorest.ParseJSON(req.Body, &s); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	s.OriginalTransactionID = strings.TrimSpace(s.OriginalTransactionID)
	if s.OriginalTransactionID == "" {
		_ = render.New(w).Unprocessable(&render.ValidationError{
			Message: "Original transaction id is required",
			Field:   "originalTransactionId",
			Code:    render.CodeInvalid,
		})
		return
	}

	snapshot, err := router.iapRepo.Unlink(s.OriginalTransactionID, readerIDs)
	if err != nil {
		if errors.Is(err, apple.ErrUnlinkMismatchedFTC) {
			_ = render.New(w).BadRequest(err.Error())
			return
		}

		_ = render.New(w).DBError(err)
		return
	}

	go func() {
		_ = router.readerRepo.BackUpMember(snapshot)
	}()

	_ = render.New(w).NoContent()
}

// UpsertSubs performs exactly the same step as VerifyReceipt.
// The two only differs in the data they send back.
func (router IAPRouter) UpsertSubs(w http.ResponseWriter, req *http.Request) {
	var input apple.ReceiptInput
	if err := gorest.ParseJSON(req.Body, &input); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}
	if ve := input.Validate(); ve != nil {
		_ = render.New(w).Unprocessable(ve)
		return
	}

	resp, resErr := router.doVerification(input.ReceiptData)
	if resErr != nil {
		_ = render.New(w).JSON(resErr.StatusCode, resErr)
		return
	}

	sub, err := resp.Subscription()

	err = router.iapRepo.UpsertSubscription(sub)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	snapshot, err := router.iapRepo.UpdateMembership(sub)
	if err != nil {
		return
	}

	// Update subscription and possible membership in background since this step is irrelevant to verification.
	if !snapshot.IsZero() {
		go func() {
			_ = router.readerRepo.BackUpMember(snapshot)
		}()
	}

	_ = render.New(w).OK(sub)
}

// RefreshSubs updates an existing apple receipt and optional associated subscription.
// Returns apple.Subscription which contains the essential
// fields to represent a user's subscription.
func (router IAPRouter) RefreshSubs(w http.ResponseWriter, req *http.Request) {
	defer logger.Sync()
	sugar := logger.Sugar()

	originalTxID, err := getURLParam(req, "id").ToString()
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	// Find existing subscription data for this original transaction id.
	// If not found, returns 404.
	sub, err := router.iapRepo.LoadSubscription(originalTxID)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	// Load the receipt file from disk.
	// If error occurred, returns 404.
	b, err := iaprepo.LoadReceipt(sub.OriginalTransactionID, sub.Environment)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).NotFound()
		return
	}

	resp, resErr := router.doVerification(string(b))
	if resErr != nil {
		sugar.Error(err)
		_ = render.New(w).JSON(resErr.StatusCode, resErr)
		return
	}

	// If err occurred, it indicates program has bugs.
	updatedSubs, err := resp.Subscription()
	if err != nil {
		_ = render.New(w).InternalServerError(err.Error())
		return
	}

	// Update subscription and possible membership in background since this step is irrelevant to verification.
	err = router.iapRepo.UpsertSubscription(updatedSubs)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	snapshot, err := router.iapRepo.UpdateMembership(sub)
	if err != nil {
		return
	}

	if !snapshot.IsZero() {
		go func() {
			_ = router.readerRepo.BackUpMember(snapshot)
		}()
	}

	_ = render.New(w).OK(updatedSubs)
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
		_ = render.New(w).NotFound()
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

// WebHook receives app store server-to-server notification.
func (router IAPRouter) WebHook(w http.ResponseWriter, req *http.Request) {
	defer logger.Sync()
	sugar := logger.Sugar()

	var wh apple.WebHook
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		sugar.Error(err)
		return
	}

	sugar.Infof("Apple webhook raw: %s", b)

	if err := json.Unmarshal(b, &wh); err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest("")
		return
	}

	_ = router.iapRepo.SaveNotification(wh.Schema())

	if !wh.UnifiedReceipt.Validate() {
		sugar.Infof("invalid webhook unified receipt")
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

	router.iapRepo.SaveResponsePayload(wh.UnifiedReceipt)

	// Build apple's subscription and save it.
	sub, err := wh.UnifiedReceipt.Subscription()
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest("")
		return
	}

	// Insert/Update subscription
	_ = router.iapRepo.UpsertSubscription(sub)

	// Update membership if exists.
	snapshot, err := router.iapRepo.UpdateMembership(sub)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	// Snapshot might be empty is this subscription is linked to ftc account yet.
	if !snapshot.IsZero() {
		go func() {
			_ = router.readerRepo.BackUpMember(snapshot)
		}()
	}

	_ = render.New(w).OK(nil)
}
