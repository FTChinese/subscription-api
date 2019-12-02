package controller

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/view"
	"github.com/jmoiron/sqlx"
	"gitlab.com/ftchinese/subscription-api/models/apple"
	"gitlab.com/ftchinese/subscription-api/models/reader"
	"gitlab.com/ftchinese/subscription-api/models/util"
	"gitlab.com/ftchinese/subscription-api/repository/iaprepo"
	"net/http"
)

type IAPRouter struct {
	iapEnv iaprepo.IAPEnv
}

func NewIAPRouter(db *sqlx.DB, c util.BuildConfig) IAPRouter {
	return IAPRouter{
		iapEnv: iaprepo.NewIAPEnv(db, c),
	}
}

// VerifyReceipt perform app store receipt verification
// Input {"receipt-data": string}
func (router IAPRouter) VerifyReceipt(w http.ResponseWriter, req *http.Request) {
	var receiptReq apple.VerificationRequestBody
	if err := gorest.ParseJSON(req.Body, &receiptReq); err != nil {
		logger.WithField("trace", "IAPRouter.VerifyReceipt").Error(err)

		_ = view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	// Validate input.
	if r := receiptReq.Validate(); r != nil {
		_ = view.Render(w, view.NewUnprocessable(r))
		return
	}

	resp, err := router.iapEnv.VerifyReceipt(receiptReq)
	if err != nil {
		_ = view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	logger.WithField("trace", "IAPRouter.VerifyReceipt").
		Infof("Environment %s, is retryable %t, status %d", resp.Environment, resp.IsRetryable, resp.Status)

	// If the response is not valid.
	if !resp.Validate() {
		logger.WithField("trace", "IAPRouter.VerifyReceipt").Info("Verification response is not valid")

		err := router.iapEnv.SaveVerificationFailure(resp.Failure(receiptReq.ReceiptData))

		if err != nil {
			logger.WithField("trace", "IAPRouter.VerifyReceipt").Error(err)
		}

		_ = view.Render(w, view.NewResponse().SetBody(map[string]bool{
			"isValid": false,
		}))
		return
	}

	// Save verification session
	r := resp.FindLatestReceipt()

	go func() {
		_ = router.iapEnv.SaveVerificationSession(resp.SessionSchema(r))
	}()

	// save latest receipt array
	go func() {
		for _, v := range resp.LatestReceiptInfo {
			_ = router.iapEnv.SaveCustomerReceipt(v.Schema(resp.Environment))
		}
	}()

	// Save pending renewal array
	go func() {
		for _, v := range resp.PendingRenewalInfo {
			_ = router.iapEnv.SavePendingRenewal(v.Schema(resp.Environment))
		}
	}()

	// Save the receipt as a token for status polling.
	receiptToken := resp.ReceiptToken(r.OriginalTransactionID)
	go func() {
		_ = router.iapEnv.SaveReceiptToken(receiptToken)
	}()

	// Create a subscription.
	// Here we do not have reader's ids
	sub := resp.Subscription(reader.MemberID{}, r)
	go func() {
		_ = router.iapEnv.CreateSubscription(sub)
	}()

	_ = view.Render(w, view.NewResponse().SetBody(map[string]bool{
		"isValid": true,
	}))
}

// WebHook receives app store server-to-server notification.
func (router IAPRouter) WebHook(w http.ResponseWriter, req *http.Request) {

}

// Link associates apple subscription to FTC reader id or wechat union id.
// User id could be either FTC id or wechat union id.
// FTC id should take precedence over wechat union id to avoid a situation that we linked membership only while account is
// not linked.
func (router IAPRouter) Link(w http.ResponseWriter, req *http.Request) {
	log := logger.WithField("trace", "IAPRouter.Link")

	// Get ftc id or union id
	memberID, err := reader.NewMemberID(
		req.Header.Get(ftcIDKey),
		req.Header.Get(unionIDKey))

	if err != nil {
		_ = view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	tx, err := router.iapEnv.BeginOrderTx()
	if err != nil {
		_ = view.Render(w, view.NewInternalError(err.Error()))
		return
	}

	// Retrieve current membership. Chances are it is empty.
	// The membership must be empty or invalid to permit proceeding.
	// Invalid means:
	// Expired for alipay and wechat pay
	// Expired for stripe pay and not auto-renewal
	// Expires for apple pay and not auto-renewal
	member, err := tx.RetrieveMember(memberID)
	if err != nil {
		log.Error(err)
		_ = tx.Rollback()
		_ = view.Render(w, view.NewDBFailure(err))
		return
	}

	if member.IsValid() {
		r := view.NewReason()
		r.Field = "membership"
		r.Code = view.CodeAlreadyExists
		r.SetMessage("Account already has a valid membership through non-apple subscription")
		_ = tx.Rollback()
		_ = view.Render(w, view.NewUnprocessable(r))
	}

	if !member.IsZero() && member.ID.IsZero() {
		member.GenerateID()
		go func() {

		}()
	}
	// Verify receipt to get the original transaction id

	// Perform CRUD in background

	// Build apple.Subscription

	// Upsert apple.Subscription

	// Build paywall.Membership from apple.Subscription.

	// Upsert paywall.Membership.

	if err := tx.Commit(); err != nil {
		_ = view.Render(w, view.NewInternalError(err.Error()))
		return
	}

	_ = view.Render(w, view.NewNoContent())
}
