package controller

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/postoffice"
	"github.com/FTChinese/go-rest/view"
	"gitlab.com/ftchinese/subscription-api/models/apple"
	"gitlab.com/ftchinese/subscription-api/models/reader"
	"gitlab.com/ftchinese/subscription-api/models/subscription"
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
				resp.Failure(receiptReq.ReceiptData),
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

func (router IAPRouter) handleVerificationResponse(resp apple.VerificationResponseBody, id reader.MemberID) apple.Subscription {

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
	sub := resp.Subscription(id, r)
	go func() {
		_ = router.iapEnv.CreateSubscription(sub)
	}()

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

	_ = router.handleVerificationResponse(result.resp, reader.MemberID{})

	_ = view.Render(w, view.NewResponse().SetBody(map[string]bool{
		"isValid": true,
	}))
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

	tx, err := router.subEnv.BeginOrderTx()
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
	currMembership, err := tx.RetrieveMember(memberID)
	if err != nil {
		log.Error(err)
		_ = tx.Rollback()
		_ = view.Render(w, view.NewDBFailure(err))
		return
	}

	// If target membership is valid, stop processing.
	if currMembership.IsValid() {
		r := view.NewReason()
		r.Field = "membership"
		r.Code = view.CodeAlreadyExists
		r.SetMessage("Account already has a valid membership through non-apple subscription")
		_ = tx.Rollback()
		_ = view.Render(w, view.NewUnprocessable(r))
	}

	// Verify receipt to get the original transaction id
	result := router.doVerification(req)

	switch {
	case result.err != nil:
		_ = tx.Rollback()
		_ = view.Render(w, view.NewBadRequest(result.err.Error()))
		return

	case result.invalidRequest != nil:
		_ = tx.Rollback()
		_ = view.Render(w, view.NewUnprocessable(result.invalidRequest))
		return

	case result.invalidResponse:
		_ = tx.Rollback()

		r := view.NewReason()
		r.Field = "receipt-data"
		r.Code = view.CodeInvalid
		r.SetMessage("receipt data is not valid")

		_ = view.Render(w, view.NewUnprocessable(r))
		return
	}

	// Here we use the member id retrieved from db to avoid
	// accidentally perform wechat-ftc linking.
	sub := router.handleVerificationResponse(result.resp, currMembership.MemberID)

	iapMembership := sub.Membership()

	// Upsert paywall.Membership.
	if currMembership.IsZero() {
		if err := tx.CreateMember(iapMembership); err != nil {
			_ = tx.Rollback()
			_ = view.Render(w, view.NewDBFailure(err))
			return
		}
	} else {
		// Take a snapshot of current membership.
		if currMembership.ID.IsZero() {
			currMembership.ID = iapMembership.ID
		}

		snapshot := subscription.NewMemberSnapshot(
			currMembership,
			enum.SnapshotReasonDelete)

		go func() {
			_ = router.subEnv.BackUpMember(snapshot)
		}()

		// Update membership.
		if err := tx.UpdateMember(iapMembership); err != nil {
			_ = tx.Rollback()
			_ = view.Render(w, view.NewDBFailure(err))
			return
		}
	}

	if err := tx.Commit(); err != nil {
		_ = view.Render(w, view.NewInternalError(err.Error()))
		return
	}

	_ = view.Render(w, view.NewNoContent())
}

// WebHook receives app store server-to-server notification.
func (router IAPRouter) WebHook(w http.ResponseWriter, req *http.Request) {
	var wh apple.WebHook
	if err := gorest.ParseJSON(req.Body, &wh); err != nil {
		_ = view.Render(w, view.NewBadRequest(""))
		return
	}

	err := router.iapEnv.SaveNotification(wh.Schema())
}
