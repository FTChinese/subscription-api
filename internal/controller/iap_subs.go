package controller

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/repository/iaprepo"
	"github.com/FTChinese/subscription-api/pkg/apple"
	"net/http"
)

func (router IAPRouter) ListSubs(w http.ResponseWriter, req *http.Request) {
	p := gorest.GetPagination(req)
	list, err := router.iapRepo.ListSubs(p)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(list)
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

func (router IAPRouter) LoadSubs(w http.ResponseWriter, req *http.Request) {
	id, err := getURLParam(req, "id").ToString()
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	s, err := router.iapRepo.LoadSubs(id)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(s)
}

// RefreshSubs updates an existing apple receipt and optional associated subscription.
// Returns apple.Subscription which contains the essential
// fields to represent a user's subscription.
func (router IAPRouter) RefreshSubs(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	originalTxID, err := getURLParam(req, "id").ToString()
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	// Find existing subscription data for this original transaction id.
	// If not found, returns 404.
	sub, err := router.iapRepo.LoadSubs(originalTxID)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	// Load the receipt file from disk.
	// If error occurred, returns 404.
	b, err := iaprepo.LoadReceipt(sub.BaseSchema)
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

	// Use old subscription's creation time.
	updatedSubs.CreatedUTC = sub.CreatedUTC

	_ = render.New(w).OK(updatedSubs)
}

// LoadReceipt retrieves the subscription data for
// an original transaction id, together with the
// receipt used to verify it.
func (router IAPRouter) LoadReceipt(w http.ResponseWriter, req *http.Request) {
	id, _ := getURLParam(req, "id").ToString()

	sub, err := router.iapRepo.LoadSubs(id)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	b, err := iaprepo.LoadReceipt(sub.BaseSchema)
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
