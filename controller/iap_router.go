package controller

import (
	"github.com/FTChinese/go-rest"
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

func (router IAPRouter) handleTransactions(ur apple.UnifiedReceipt, originalTransactionID string) {

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
	receiptToken := ur.ReceiptToken(
		originalTransactionID,
	)
	go func() {
		_ = router.iapEnv.SaveReceiptToken(receiptToken)
	}()
}

// VerifyReceipt perform app store receipt verification
// Input {"receipt-data": string}
func (router IAPRouter) VerifyReceipt(w http.ResponseWriter, req *http.Request) {
	log := logger.WithField("trace", "IAPRouter.VerifyReceipt")

	// Parse input data.
	var receiptReq apple.VerificationRequestBody
	if err := gorest.ParseJSON(req.Body, &receiptReq); err != nil {
		log.Error(err)
		_ = view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	// Validate input.
	if r := receiptReq.Validate(); r != nil {
		_ = view.Render(w, view.NewUnprocessable(r))
		return
	}

	// Verify
	resp, err := router.iapEnv.VerifyReceipt(receiptReq)
	if err != nil {
		_ = view.Render(w, view.NewBadRequest(err.Error()))
		return
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

		r := view.NewReason()
		r.Field = "receipt-data"
		r.Code = view.CodeInvalid
		r.SetMessage("receipt data is not valid")
		_ = view.Render(w, view.NewUnprocessable(r))
		return
	}

	// Save verification session
	transaction := resp.FindLatestTransaction()
	go func() {
		_ = router.iapEnv.SaveVerificationSession(
			resp.SessionSchema(transaction),
		)
	}()

	sub := resp.Subscription(transaction)
	_ = router.iapEnv.CreateSubscription(sub)

	// Check user's ids.
	memberID, err := reader.NewMemberID(
		req.Header.Get(ftcIDKey),
		req.Header.Get(unionIDKey),
	)

	// If reader ids not found, it indicates this is not
	// used to link membership.
	if err != nil {
		_ = view.Render(w, view.NewResponse().SetBody(map[string]bool{
			"isValid": true,
		}))
		return
	}

	// Start to link apple subscription to ftc membership.
	err = router.iapEnv.Link(sub, memberID)

	switch err {

	case subscription.ErrValidFTCMember:
		r := view.NewReason()
		r.Field = "link"
		r.Code = "target_valid"
		r.SetMessage("Cannot linking iap subscription to your logged in account since your ftc account already has a valid membership.")
		_ = view.Render(w, view.NewUnprocessable(r))
		return

	case nil:
		_ = view.Render(w, view.NewNoContent())
		return

	default:
		_ = view.Render(w, view.NewDBFailure(err))
		return
	}
}

// WebHook receives app store server-to-server notification.
func (router IAPRouter) WebHook(w http.ResponseWriter, req *http.Request) {
	var wh apple.WebHook
	if err := gorest.ParseJSON(req.Body, &wh); err != nil {
		logger.WithField("trace", "IAPRouter.WebHook").Error(err)

		_ = view.Render(w, view.NewBadRequest(""))
		return
	}

	_ = router.iapEnv.SaveNotification(wh.Schema())

	// Find the latest transaction and save transaction
	// history.
	transaction := wh.UnifiedReceipt.FindLatestTransaction()

	router.handleTransactions(
		wh.UnifiedReceipt,
		transaction.OriginalTransactionID,
	)

	// Build apple's subscription and save it.
	sub := wh.UnifiedReceipt.Subscription(transaction)
	if err := router.iapEnv.CreateSubscription(sub); err != nil {
		_ = view.Render(w, view.NewBadRequest(""))
	}

	// Retrieve apple subscription by original transaction id.
	// together with membership.
	// If apple subscription is not linked, stop.
	// If apple subscription is linked, update the membership,
	// continue to find the membership by linked id,
	// and take a snapshot of membership, and then
	// update it.

	tx, err := router.iapEnv.BeginTx(sub.Environment)
	if err != nil {
		_ = view.Render(w, view.NewDBFailure(err))
		return
	}

	currMember, err := tx.RetrieveAppleMember(sub.OriginalTransactionID)
	if err != nil {
		_ = tx.Rollback()
		_ = view.Render(w, view.NewBadRequest(""))
		return
	}

	if currMember.IsZero() {
		_ = tx.Commit()
		_ = view.Render(w, view.NewResponse())
		return
	}

	newMember := sub.NewMembership(currMember.MemberID)
	newMember.ID = currMember.ID

	// Take a snapshot.
	go func() {
		_ = router.iapEnv.BackUpMember(
			subscription.NewMemberSnapshot(
				currMember,
				enum.SnapshotReasonAppleIAP,
			),
			sub.Environment,
		)
	}()

	if err := tx.UpdateMember(newMember); err != nil {
		_ = tx.Rollback()
		_ = view.Render(w, view.NewBadRequest(""))
		return
	}

	_ = tx.Commit()

	_ = view.Render(w, view.NewResponse())
}
