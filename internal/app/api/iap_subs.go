package api

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/pkg/apple"
	"github.com/FTChinese/subscription-api/pkg/xhttp"
	"net/http"
)

// ListSubs gets a list of a user's apple subscription.
//
// GET /apple/subs?ftc_id=<uuid>
func (router IAPRouter) ListSubs(w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	ftcID := xhttp.GetFtcID(req.Header)
	p := gorest.GetPagination(req)

	list, err := router.Repo.ListSubs(ftcID, p)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(list)
}

// UpsertSubs performs exactly the same step as VerifyReceipt.
// The two only differs in the data they send back.
//
// Input:
// receiptData: string;
func (router IAPRouter) UpsertSubs(w http.ResponseWriter, req *http.Request) {
	defer router.Logger.Sync()
	sugar := router.Logger.Sugar()

	var input apple.ReceiptInput
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

	resp, resErr := router.doVerification(input.ReceiptData)
	if resErr != nil {
		sugar.Error(resErr)
		_ = render.New(w).JSON(resErr.StatusCode, resErr)
		return
	}

	sub, err := apple.NewSubscription(resp.UnifiedReceipt)

	result, err := router.Repo.SaveSubs(sub)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	if !result.Versioned.IsZero() {
		go func() {
			err := router.ReaderRepo.VersionMembership(result.Versioned)
			if err != nil {
				sugar.Error(err)
			}
		}()
	}

	_ = render.New(w).OK(sub)
}

// LoadSubs retrieves apple.Subscription for the specified
// original transaction id.
//
// GET /apple/subs/{id}
func (router IAPRouter) LoadSubs(w http.ResponseWriter, req *http.Request) {
	id, err := xhttp.GetURLParam(req, "id").ToString()
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	s, err := router.Repo.LoadSubs(id)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(s)
}

// RefreshSubs updates an existing apple receipt and optional associated membership.
// Returns apple.Subscription which contains the essential
// fields to represent a user's subscription.
//
// PATCH /apple/subs/{id}
func (router IAPRouter) RefreshSubs(w http.ResponseWriter, req *http.Request) {
	defer router.Logger.Sync()
	sugar := router.Logger.Sugar()

	origTxID, err := xhttp.GetURLParam(req, "id").ToString()
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	// Find existing subscription data for this original transaction id.
	// If not found, returns 404.
	sub, err := router.Repo.LoadSubs(origTxID)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	// Load the receipt file from disk.
	// If error occurred, returns 404.
	receipt, err := router.Repo.LoadReceipt(sub.BaseSchema, false)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).NotFound("Not found the Apple subscription to refresh")
		return
	}

	resp, resErr := router.doVerification(receipt)
	if resErr != nil {
		sugar.Error(err)
		_ = render.New(w).JSON(resErr.StatusCode, resErr)
		return
	}

	// If err occurred, it indicates program has bugs.
	updatedSubs, err := apple.NewSubscription(resp.UnifiedReceipt)
	if err != nil {
		_ = render.New(w).InternalServerError(err.Error())
		return
	}
	// Use old subscription's creation time.
	updatedSubs.CreatedUTC = sub.CreatedUTC

	// Update subscription and possible membership
	result, err := router.Repo.SaveSubs(updatedSubs)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	if !result.Versioned.IsZero() {
		go func() {
			err := router.ReaderRepo.VersionMembership(result.Versioned)
			if err != nil {
				sugar.Error(err)
			}
		}()
	}

	_ = render.New(w).OK(result)
}
