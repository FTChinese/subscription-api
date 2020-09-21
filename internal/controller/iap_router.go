package controller

import (
	"encoding/json"
	"errors"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"

	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/postoffice"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/repository/iaprepo"
	"github.com/FTChinese/subscription-api/internal/repository/readerrepo"
	"github.com/FTChinese/subscription-api/pkg/apple"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/pkg/letter"
	"github.com/jmoiron/sqlx"
)

type IAPRouter struct {
	iapRepo    iaprepo.Env
	readerRepo readerrepo.Env
	postman    postoffice.PostOffice
	rdb        *redis.Client

	secret    string
	iapClient iaprepo.Client
	logger    *zap.Logger
}

func NewIAPRouter(db *sqlx.DB, p postoffice.PostOffice, cfg config.BuildConfig, logger *zap.Logger) IAPRouter {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.MustRedisAddr(),
		Password: "",
		DB:       0,
	})

	return IAPRouter{
		iapRepo:    iaprepo.NewEnv(db, rdb, logger),
		readerRepo: readerrepo.NewEnv(db),
		postman:    p,

		secret:    config.MustIAPSecret(),
		iapClient: iaprepo.NewClient(cfg.Sandbox()),
		logger:    logger,
	}
}

// Verify a receipt and get response from app store.
// This the shared action when verifying receipt,
// link account and unlink account.
// See https://developer.apple.com/documentation/storekit/in-app_purchase/validating_receipts_with_the_app_store
func (router IAPRouter) doVerification(receipt string) (apple.VerificationResp, *render.ResponseError) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

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
	if ve := resp.Validate(); ve != nil {
		sugar.Info("IAP verification response is not valid")
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
// receiptData: string
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

			snapshot, err := router.iapRepo.SaveSubs(sub)
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

// Link links IAP subscription to FTC account.
// This step does not perform verification.
// It only links an existing subscription to ftc account.
// You should ask the /subscription/* endpoint to
// update data and get the original transaction id.
//
// Input:
// ftcId: string;
// originalTxId: string;
//
// Response: the linked Membership.
func (router IAPRouter) Link(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	var input apple.LinkInput
	if err := gorest.ParseJSON(req.Body, &input); err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}
	if ve := input.Validate(); ve != nil {
		sugar.Error(ve)
		_ = render.New(w).Unprocessable(ve)
		return
	}

	ftcAccount, err := router.readerRepo.FtcAccountByFtcID(input.FtcID)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	iapSubs, err := router.iapRepo.LoadSubs(input.OriginalTxID)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	// Start to link apple subscription to ftc membership.
	linkResult, err := router.iapRepo.Link(ftcAccount, iapSubs)

	if err != nil {
		sugar.Error(err)
		var ve *render.ValidationError
		// ValidationError indicates the link is not allowed.
		if errors.As(err, &ve) {
			_ = render.New(w).Unprocessable(ve)
			return
		}

		_ = render.New(w).DBError(err)
		return
	}

	if !linkResult.Snapshot.IsZero() {
		go func() {
			_ = router.readerRepo.BackUpMember(linkResult.Snapshot)
		}()
	}

	// Send notification email if this is initial link.
	if linkResult.Initial {
		sugar.Info("Initial link. Sending email....")
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
// ftcId: string;
// originalTxId: string;
func (router IAPRouter) Unlink(w http.ResponseWriter, req *http.Request) {

	var input apple.LinkInput
	// 400 Bad Request if request body cannot be parsed.
	if err := gorest.ParseJSON(req.Body, &input); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	// 422 Unprocessable for request body is not valid.
	if ve := input.Validate(); ve != nil {
		_ = render.New(w).Unprocessable(ve)
	}

	// This will retrieve membership by apple original transaction id.
	// So if target does not exists, if will simply gives 404 error.
	snapshot, err := router.iapRepo.Unlink(input)
	if err != nil {
		var ve *render.ValidationError
		if errors.As(err, &ve) {
			_ = render.New(w).Unprocessable(ve)
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

// WebHook receives app store server-to-server notification.
func (router IAPRouter) WebHook(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

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
	// TODO: use original transaction id to search ftc_vip_ios table;
	// if found, use the associated vip_id (where vip_id_alias is NULL) to find membership in ftc_vip table;
	// if this membership payMethod is null, and expireDate is not after sub.ExpireDateUTC,
	// then we should update this membership using this subscription.
	// This approach can be used in webhook notification and verify-receipt.
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
