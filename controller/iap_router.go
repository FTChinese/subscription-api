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
	env iaprepo.IAPEnv
}

func NewIAPRouter(db *sqlx.DB, c util.BuildConfig) IAPRouter {
	return IAPRouter{
		env: iaprepo.NewIAPEnv(db, c),
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

	resp, err := router.env.VerifyReceipt(receiptReq)
	if err != nil {
		_ = view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	logger.WithField("trace", "IAPRouter.VerifyReceipt").
		Infof("Environment %s, is retryable %t, status %d", resp.Environment, resp.IsRetryable, resp.Status)

	// If the response is not valid.
	if !resp.Validate() {
		logger.WithField("trace", "IAPRouter.VerifyReceipt").Info("Verification response is not valid")

		err := router.env.SaveVerificationFailure(resp.Failure(receiptReq.ReceiptData))

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
		_ = router.env.SaveVerificationSession(resp.SessionSchema(r))
	}()

	// save latest receipt array
	go func() {
		for _, v := range resp.LatestReceiptInfo {
			_ = router.env.SaveCustomerReceipt(v.Schema(resp.Environment))
		}
	}()

	// Save pending renewal array
	go func() {
		for _, v := range resp.PendingRenewalInfo {
			_ = router.env.SavePendingRenewal(v.Schema(resp.Environment))
		}
	}()

	// Save the receipt as a token for status polling.
	receiptToken := resp.ReceiptToken(r.OriginalTransactionID)
	go func() {
		_ = router.env.SaveReceiptToken(receiptToken)
	}()

	// Create a subscription.
	// Here we do not have reader's ids
	sub := resp.Subscription(reader.MemberID{}, r)
	go func() {
		_ = router.env.CreateSubscription(sub)
	}()

	_ = view.Render(w, view.NewResponse().SetBody(map[string]bool{
		"isValid": true,
	}))
}

// WebHook receives app store server-to-server notification.
func (router IAPRouter) WebHook(w http.ResponseWriter, req *http.Request) {

}

// Link associates apple subscription to FTC reader id or wechat union id.
func (router IAPRouter) Link(w http.ResponseWriter, req *http.Request) {
	// Get ftc id or union id

	// Retrieve current membership. Chances are it is empty.
	// The membership must be empty or invalid to permit proceeding.
	// Invalid means:
	// Expired for alipay and wechat pay
	// Expired for stripe pay and not auto-renewal
	// Expires for apple pay and not auto-renewal

	// Verify receipt to get the original transaction id

	// Perform CRUD in background

	// Build apple.Subscription

	// Upsert apple.Subscription

	// Build paywall.Membership from apple.Subscription.

	// Upsert paywall.Membership.
}
