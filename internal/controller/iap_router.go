package controller

import (
	"encoding/json"
	"errors"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/postman"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"

	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/repository/iaprepo"
	"github.com/FTChinese/subscription-api/pkg/apple"
)

type IAPRouter struct {
	iapRepo iaprepo.Env
	postman postman.Postman

	isLive    bool
	iapClient iaprepo.Client
	logger    *zap.Logger
}

func NewIAPRouter(
	dbs db.ReadWriteMyDBs,
	rdb *redis.Client,
	logger *zap.Logger,
	p postman.Postman,
	isLive bool,
) IAPRouter {

	return IAPRouter{
		iapRepo:   iaprepo.NewEnv(dbs, rdb, logger),
		postman:   p,
		isLive:    isLive,
		iapClient: iaprepo.NewClient(logger),
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

	// Send data to IAP endpoint for verification
	resp, err := router.iapClient.VerifyAndValidate(receipt, router.isLive)

	if err != nil {
		sugar.Error(err)

		var ve *render.ValidationError
		if errors.As(err, &ve) {
			return apple.VerificationResp{}, render.NewUnprocessable(ve)
		}

		return apple.VerificationResp{}, render.NewBadRequest(err.Error())
	}

	// Save the decoded receipt as a session of verification
	go func() {
		_ = router.iapRepo.SaveDecodedReceipt(
			resp.ReceiptSchema(),
		)
	}()

	// Dissect and save other fields in the verification response.
	router.iapRepo.SaveUnifiedReceipt(resp.UnifiedReceipt)
	return resp, nil
}

func (router IAPRouter) processSubsResult(snapshot reader.MemberSnapshot) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	// Backup previous membership
	if !snapshot.IsZero() {
		err := router.iapRepo.ArchiveMember(snapshot)
		if err != nil {
			sugar.Error(err)
		}
	}
}

// VerifyReceipt verifies if the receipt data send by client is valid. After app store responded,
// its latest_receipt, latest_receipt_info, pending_renewal_info are saved in DB in background thread.
// An apple.Subscription is created from the response, which is saved or updated if already exists,
// and then user's membership is updated if it exists.
//
// Input
// receiptData: string
func (router IAPRouter) VerifyReceipt(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	// Parse request body.
	var input apple.ReceiptInput
	if err := gorest.ParseJSON(req.Body, &input); err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}
	// Validate input data
	if ve := input.Validate(); ve != nil {
		sugar.Error(ve)
		_ = render.New(w).Unprocessable(ve)
		return
	}

	// Perform verification.
	resp, resErr := router.doVerification(input.ReceiptData)
	if resErr != nil {
		sugar.Error(resErr)
		_ = render.New(w).JSON(resErr.StatusCode, resErr)
		return
	}

	// Create apple.Subscription.
	// TODO: this subscription does not know if it is linked to an email.
	sub, err := apple.NewSubscription(resp.UnifiedReceipt)
	if err != nil {
		sugar.Error(err)
	}

	// Update subscription and possible membership in background since this step is irrelevant to verification.
	if err == nil {
		go func() {

			result, err := router.iapRepo.SaveSubs(sub)
			if err != nil {
				sugar.Error(err)
				return
			}

			router.processSubsResult(result.Snapshot)
		}()
	}

	_ = render.New(w).OK(resp)
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

	err = router.iapRepo.SaveWebhook(apple.NewWebHookSchema(wh))
	if err != nil {
		sugar.Error(err)
	}

	if !wh.UnifiedReceipt.IsSubscribed() {
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

	router.iapRepo.SaveUnifiedReceipt(wh.UnifiedReceipt)

	// Build apple's subscription and save it.
	sub, err := apple.NewSubscription(wh.UnifiedReceipt)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest("")
		return
	}

	// Update membership if exists.
	// if found, use the associated vip_id (where vip_id_alias is NULL) to find membership in ftc_vip table;
	// if this membership payMethod is null, and expireDate is not after sub.ExpireDateUTC,
	// then we should update this membership using this subscription.
	// This approach can be used in webhook notification and verify-receipt.
	result, err := router.iapRepo.SaveSubs(sub)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	// Snapshot might be empty is this subscription is linked to ftc account yet.
	go func() {
		router.processSubsResult(result.Snapshot)
	}()

	_ = render.New(w).OK(nil)
}
