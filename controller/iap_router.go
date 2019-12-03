package controller

import (
	"github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/postoffice"
	"github.com/FTChinese/go-rest/view"
	"gitlab.com/ftchinese/subscription-api/models/apple"
	"gitlab.com/ftchinese/subscription-api/models/reader"
	"gitlab.com/ftchinese/subscription-api/repository/iaprepo"
	"gitlab.com/ftchinese/subscription-api/repository/subrepo"
	"net/http"
)

type IAPRouter struct {
	PayRouter
	iapEnv iaprepo.IAPEnv
}

func NewIAPRouter(iapEnv iaprepo.IAPEnv, subEnv subrepo.SubEnv, p postoffice.Postman) IAPRouter {
	return IAPRouter{
		PayRouter: PayRouter{
			subEnv:  subEnv,
			postman: p,
		},
		iapEnv: iapEnv,
	}
}

type verificationResult struct {
	resp            apple.VerificationResponseBody
	err             error
	invalidRequest  *view.Reason
	invalidResponse bool
}

func (router IAPRouter) doVerification(req *http.Request) verificationResult {

	log := logger.WithField("trace", "IAPRouter.doVerification")

	var receiptReq apple.VerificationRequestBody
	if err := gorest.ParseJSON(req.Body, &receiptReq); err != nil {
		log.Error(err)

		return verificationResult{
			resp:            apple.VerificationResponseBody{},
			err:             err,
			invalidRequest:  nil,
			invalidResponse: false,
		}
	}

	// Validate input.
	if r := receiptReq.Validate(); r != nil {
		return verificationResult{
			resp:            apple.VerificationResponseBody{},
			err:             nil,
			invalidRequest:  r,
			invalidResponse: false,
		}
	}

	// Verify
	resp, err := router.iapEnv.VerifyReceipt(receiptReq)
	if err != nil {
		return verificationResult{
			resp:            apple.VerificationResponseBody{},
			err:             err,
			invalidRequest:  nil,
			invalidResponse: false,
		}
	}

	log.Infof("Environment %s, is retryable %t, status %d", resp.Environment, resp.IsRetryable, resp.Status)

	// If the response is not valid.
	if !resp.Validate() {
		log.Info("Verification response is not valid")

		go func() {
			_ = router.iapEnv.SaveVerificationFailure(
				resp.FailureSchema(receiptReq.ReceiptData),
			)
		}()

		return verificationResult{
			resp:            apple.VerificationResponseBody{},
			err:             nil,
			invalidRequest:  nil,
			invalidResponse: true,
		}
	}

	return verificationResult{
		resp:            resp,
		err:             nil,
		invalidRequest:  nil,
		invalidResponse: false,
	}
}

func (router IAPRouter) handleTransactions(ur apple.UnifiedReceipt, id reader.MemberID) apple.Subscription {

	// Save verification session
	transaction := ur.FindLatestTransaction()

	// TODO: how to save verification session?
	//go func() {
	//	_ = router.iapEnv.SaveVerificationSession(
	//		resp.SessionSchema(transaction),
	//	)
	//}()

	// save latest receipt array
	go func() {
		for _, v := range ur.LatestTransactions {
			_ = router.iapEnv.SaveCustomerReceipt(
				v.Schema(ur.Environment),
			)
		}
	}()

	// Save pending renewal array
	go func() {
		for _, v := range ur.PendingRenewalInfo {
			_ = router.iapEnv.SavePendingRenewal(
				v.Schema(ur.Environment),
			)
		}
	}()

	// Save the receipt as a token for status polling.
	receiptToken := ur.ReceiptToken(transaction.OriginalTransactionID)
	go func() {
		_ = router.iapEnv.SaveReceiptToken(receiptToken)
	}()

	// Create a subscription.
	// Here we do not have reader's ids
	sub := ur.Subscription(id, transaction)
	// TODO: CreateSubscription in main thread.
	//go func() {
	//	_ = router.iapEnv.CreateSubscription(sub)
	//}()

	return sub
}

// VerifyReceipt perform app store receipt verification
// Input {"receipt-data": string}
func (router IAPRouter) VerifyReceipt(w http.ResponseWriter, req *http.Request) {
	result := router.doVerification(req)

	switch {
	case result.err != nil:
		_ = view.Render(w, view.NewBadRequest(result.err.Error()))
		return

	case result.invalidRequest != nil:
		_ = view.Render(w, view.NewUnprocessable(result.invalidRequest))
		return

	case result.invalidResponse:
		_ = view.Render(w, view.NewResponse().SetBody(map[string]bool{
			"isValid": false,
		}))
		return
	}

	_ = router.handleTransactions(
		result.resp.UnifiedReceipt,
		reader.MemberID{},
	)

	_ = view.Render(w, view.NewResponse().SetBody(map[string]bool{
		"isValid": true,
	}))
}

// WebHook receives app store server-to-server notification.
func (router IAPRouter) WebHook(w http.ResponseWriter, req *http.Request) {
	var wh apple.WebHook
	if err := gorest.ParseJSON(req.Body, &wh); err != nil {
		_ = view.Render(w, view.NewBadRequest(""))
		return
	}

	_ = router.iapEnv.SaveNotification(wh.Schema())

	// Upsert each transaction
	// Upsert each pending renewal
	// Upsert verification token.
	// Upsert apple subscription.
	//sub := router.handleTransactions(wh.UnifiedReceipt, reader.MemberID{})

	// Retrieve apple subscription by original transaction id.
	// together with membership.
	// If apple subscription is not linked, stop.
	// If apple subscription is linked, update the membership,
	// continue to find the membership by linked id,
	// and take a snapshot of membership, and then
	// update it.
}
