package controller

import (
	"github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/postoffice"
	"github.com/FTChinese/go-rest/view"
	"gitlab.com/ftchinese/subscription-api/models/apple"
	"gitlab.com/ftchinese/subscription-api/models/letter"
	"gitlab.com/ftchinese/subscription-api/models/reader"
	"gitlab.com/ftchinese/subscription-api/models/subscription"
	"gitlab.com/ftchinese/subscription-api/repository/iaprepo"
	"gitlab.com/ftchinese/subscription-api/repository/rederrepo"
	"net/http"
)

type IAPRouter struct {
	iapEnv    iaprepo.IAPEnv
	readerEnv rederrepo.ReaderEnv
	postman   postoffice.Postman
}

func NewIAPRouter(iapEnv iaprepo.IAPEnv, p postoffice.Postman) IAPRouter {
	return IAPRouter{
		iapEnv:  iapEnv,
		postman: p,
	}
}

func (router IAPRouter) doVerification(w http.ResponseWriter, req *http.Request) (apple.Subscription, bool) {
	log := logger.WithField("trace", "IAPRouter.VerifyReceipt")

	// Parse input data.
	var receiptReq apple.VerificationRequestBody
	if err := gorest.ParseJSON(req.Body, &receiptReq); err != nil {
		log.Error(err)
		_ = view.Render(w, view.NewBadRequest(err.Error()))
		return apple.Subscription{}, false
	}

	// Validate input.
	if r := receiptReq.Validate(); r != nil {
		_ = view.Render(w, view.NewUnprocessable(r))
		return apple.Subscription{}, false
	}

	// Verify
	resp, err := router.iapEnv.VerifyReceipt(receiptReq)
	if err != nil {
		_ = view.Render(w, view.NewBadRequest(err.Error()))
		return apple.Subscription{}, false
	}

	log.Infof("Environment %s, is retryable %t, status %d", resp.Environment, resp.IsRetryable, resp.Status)

	// If the response is not valid.
	if !resp.Validate() {
		log.Info("Verification response is not valid")

		log.WithField("environment", resp.Environment).
			WithField("status", resp.Status).
			WithField("message", resp.GetStatusMessage()).
			WithField("receiptData", receiptReq.ReceiptData).
			Info("IAP verification failed")

		r := view.NewReason()
		r.Field = "receipt-data"
		r.Code = view.CodeInvalid
		r.SetMessage("receipt data is not valid")
		_ = view.Render(w, view.NewUnprocessable(r))
		return apple.Subscription{}, false
	}

	// Save verification session
	//transaction := resp.findLatestTransaction()

	resp.Parse()

	// Save this verification session.
	go func() {
		_ = router.iapEnv.SaveVerificationSession(
			resp.SessionSchema(),
		)
	}()

	router.saveReceiptData(resp.UnifiedReceipt)

	sub := resp.Subscription()
	_ = router.iapEnv.CreateSubscription(sub)

	return sub, true
}

func (router IAPRouter) saveReceiptData(ur apple.UnifiedReceipt) {

	// save latest receipt array
	go func() {
		for _, v := range ur.LatestTransactions {
			_ = router.iapEnv.SaveTransaction(
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
	receiptToken := ur.ReceiptToken()

	go func() {
		_ = router.iapEnv.SaveReceiptToken(receiptToken)
	}()
}

// VerifyReceipt perform app store receipt verification
// Input {"receipt-data": string}
func (router IAPRouter) VerifyReceipt(w http.ResponseWriter, req *http.Request) {

	sub, ok := router.doVerification(w, req)
	if !ok {
		return
	}

	// Check user's ids.
	memberID, err := reader.NewMemberID(
		req.Header.Get(ftcIDKey),
		req.Header.Get(unionIDKey),
	)

	// If reader ids not found, it indicates this is not
	// used to link membership.
	if err != nil {
		_ = view.Render(w, view.NewNoContent())
		return
	}

	// Start to link apple subscription to ftc membership.
	m, isNewLink, err := router.iapEnv.Link(sub, memberID)

	if err != nil {
		switch err {
		case subscription.ErrLinkToMultipleFTC:
			r := view.NewReason()
			r.Field = "link_source"
			r.Code = "already_linked"
			r.SetMessage(err.Error())
			_ = view.Render(w, view.NewUnprocessable(r))
			return

		case subscription.ErrTargetLinkedToOtherIAP:
			r := view.NewReason()
			r.Field = "link_target"
			r.Code = "already_linked"
			r.SetMessage(err.Error())
			_ = view.Render(w, view.NewUnprocessable(r))
			return

		case subscription.ErrHasValidNonIAPMember:
			r := view.NewReason()
			r.Field = "link_target"
			r.Code = "valid_non_iap"
			r.SetMessage(err.Error())
			_ = view.Render(w, view.NewUnprocessable(r))
			return

		default:
			_ = view.Render(w, view.NewDBFailure(err))
			return
		}
	}

	if isNewLink {
		go func() {
			_ = router.sendLinkedLetter(m)
		}()
	}

	_ = view.Render(w, view.NewResponse().SetBody(m))
}

// Unlink removes apple subscription id from a user's membership
func (router IAPRouter) Unlink(w http.ResponseWriter, req *http.Request) {
	sub, ok := router.doVerification(w, req)
	if !ok {
		return
	}

	if err := router.iapEnv.Unlink(sub); err != nil {
		_ = view.Render(w, view.NewDBFailure(err))
		return
	}

	_ = view.Render(w, view.NewNoContent())
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
	//transaction := wh.UnifiedReceipt.findLatestTransaction()

	wh.UnifiedReceipt.Parse()

	router.saveReceiptData(wh.UnifiedReceipt)

	// Build apple's subscription and save it.
	sub := wh.UnifiedReceipt.Subscription()
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

	tx, err := router.iapEnv.BeginTx()
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

	// Take a snapshot.
	go func() {
		_ = router.iapEnv.BackUpMember(
			currMember.Snapshot(enum.SnapshotReasonAppleLink),
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

func (router IAPRouter) sendLinkedLetter(m subscription.Membership) error {
	if m.FtcID.IsZero() {
		logger.
			WithField("trace", "IAPRouter.sendLinkedLetter").
			Info("not an ftc account")

		return nil
	}

	account, err := router.readerEnv.FindAccountByFtcID(m.FtcID.String)
	if err != nil {
		return err
	}

	parcel, err := letter.NewIAPLinkParcel(account, m)
	if err != nil {
		return err
	}

	err = router.postman.Deliver(parcel)
	if err != nil {
		return err
	}

	return nil
}
