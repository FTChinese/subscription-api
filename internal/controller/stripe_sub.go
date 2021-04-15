package controller

import (
	"github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/stripe"
	"net/http"
)

// CreateSubs create a stripe subscription.
// Input:
// tier: string;
// cycle: string;
// coupon?: string;
// defaultPaymentMethod?: string;
// idempotency?: string;
// Why this field?
//
// PITFALLS:
// If you creates a plan in CNY, and a customer is subscribed to
// it, and after that you created another plan in GBP, then
// Stripe will decline your subsequent subscription request.
// It's better to create different plans in the same currency.
// I guess Stripe takes plans in different currencies as the
// same one to avoid customer subscribing to the same plan
// in different countries and regions.
// {
// "status":400,
// "message":"You cannot combine currencies on a single customer. This customer has had a subscription, coupon, or invoice item with currency cny",
// "request_id":"req_fa0rfmytgnI22E",
// "type":"invalid_request_error"
// }
// in case user already linked wechat.
// Notification email is sent upon webhook receiving data, not here.
func (router StripeRouter) CreateSubs(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	// Get FTC id. Its presence is already checked by middleware.
	ftcID := req.Header.Get(userIDKey)
	var input stripe.SubsInput
	if err := gorest.ParseJSON(req.Body, &input); err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}
	// Validate input data.
	if err := input.Validate(); err != nil {
		sugar.Error(err)
		_ = render.New(w).Unprocessable(err)
		return
	}

	acnt, err := router.stripeRepo.BaseAccountByUUID(ftcID)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}
	// If this user is not a stripe customer yet.
	if acnt.StripeID.IsZero() {
		_ = render.New(w).NotFound("Stripe customer not found")
		return
	}

	if denied := acnt.ValidateEnv(router.config.Live()); denied != "" {
		_ = render.New(w).Forbidden(denied)
		return
	}

	sp, err := price.StripeEditions.FindByEdition(input.Edition, router.config.Live())
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	// Create stripe subscription.
	result, err := router.stripeRepo.CreateSubscription(stripe.SubsParams{
		Account:      acnt,
		Edition:      sp,
		SharedParams: input.SharedParams,
	})

	if err != nil {
		sugar.Error(err)

		err := handleErrResp(w, err)
		if err == nil {
			return
		}

		_ = render.New(w).DBError(err)
		return
	}

	// Save ftc id to stripe subscription id mapping.
	// Backup previous membership if exists.
	go func() {
		router.handleSubsResult(result)
	}()

	if result.MissingPaymentIntent {
		_ = render.New(w).BadRequest("PaymentIntent is not expanded")
	}

	_ = render.New(w).OK(result)
}

// UpdateSubs updates a stripe subscription:
// user could switch cycle of the same tier, or upgrading to a higher tier.
// Input:
// tier: string;
// cycle: string;
// coupon?: "",
// defaultPaymentMethod?: ""
// idempotency?: string
//
// Error response:
// 404 if membership if not found.
// NOTE: when updating a stripe subscription, the return payload
// `items` field contains more than one items:
// one is standard and another if premium.
// So we cannot rely on this field to find FTC plan.
func (router StripeRouter) UpdateSubs(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	// Get FTC id. Its presence is already checked by middleware.
	ftcID := req.Header.Get(userIDKey)
	var input stripe.SubsInput
	if err := gorest.ParseJSON(req.Body, &input); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}
	// Validating if user is updating to the same price is postponed to
	// building checkout intent.
	if ve := input.Validate(); ve != nil {
		_ = render.New(w).Unprocessable(ve)
		return
	}

	account, err := router.stripeRepo.BaseAccountByUUID(ftcID)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}
	// If this user is not a stripe customer yet.
	if account.StripeID.IsZero() {
		_ = render.New(w).NotFound("Stripe customer not found")
		return
	}

	if denied := account.ValidateEnv(router.config.Live()); denied != "" {
		_ = render.New(w).Forbidden(denied)
		return
	}

	sp, err := price.StripeEditions.FindByEdition(input.Edition, router.config.Live())
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	result, err := router.stripeRepo.UpdateSubscription(stripe.SubsParams{
		Account:      account,
		Edition:      sp,
		SharedParams: input.SharedParams,
	})

	if err != nil {
		sugar.Error(err)
		err := handleErrResp(w, err)
		if err == nil {
			return
		}

		_ = render.New(w).DBError(err)
		return
	}

	// Remember uuid to stripe subscription mapping;
	// Backup previous membership.
	go func() {
		router.handleSubsResult(result)
	}()

	if result.MissingPaymentIntent {
		_ = render.New(w).BadRequest("PaymentIntent not expanded")
	}

	_ = render.New(w).OK(result)
}

func (router StripeRouter) ListSubs(w http.ResponseWriter, req *http.Request) {
	// TODO: implementation
}

func (router StripeRouter) LoadSubs(w http.ResponseWriter, req *http.Request) {

	id, err := getURLParam(req, "id").ToString()
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	s, err := router.stripeRepo.RetrieveSubs(id)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(s)
}

// RefreshSubs get the latest data of a subscription if user manually requested it.
func (router StripeRouter) RefreshSubs(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	// Get the subscription id from url
	subsID, err := getURLParam(req, "id").ToString()
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	// Use Stripe SDK to retrieve data.
	ss, err := router.client.GetSubs(subsID)
	if err != nil {
		sugar.Error(err)
		err = handleErrResp(w, err)
		if err == nil {
			return
		}

		_ = render.New(w).BadRequest(err.Error())
		return
	}

	// Use Stripe customer id to find user account.
	account, err := router.stripeRepo.BaseAccountByStripeID(ss.Customer.ID)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).NotFound("Stripe customer not found")
		return
	}

	result, err := router.stripeRepo.OnSubscription(ss, stripe.SubsResultParams{
		UserIDs: account.CompoundIDs(),
		Action:  reader.ActionRefresh,
	})

	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	// Only update subs and snapshot if actually modified.
	go func() {
		router.handleSubsResult(result)
	}()

	_ = render.New(w).OK(result)
}

// CancelSubs cancels a stripe subscription at period end.
// See https://stripe.com/docs/billing/subscriptions/cancel
func (router StripeRouter) CancelSubs(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	ftcID := req.Header.Get(userIDKey)

	subsID, err := getURLParam(req, "id").ToString()
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	result, err := router.stripeRepo.CancelSubscription(stripe.CancelParams{
		FtcID:  ftcID,
		SubID:  subsID,
		Cancel: true,
	})

	if err != nil {
		sugar.Error(err)
		err := handleErrResp(w, err)
		if err == nil {
			return
		}

		_ = render.New(w).DBError(err)
		return
	}

	// Remember uuid to stripe subscription mapping;
	// Backup previous membership.
	go func() {
		router.handleSubsResult(result)
	}()

	_ = render.New(w).OK(result)
}

// ReactivateSubscription undo subscription cancellation before period ends.
func (router StripeRouter) ReactivateSubscription(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	ftcID := req.Header.Get(userIDKey)

	subsID, err := getURLParam(req, "id").ToString()
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	result, err := router.stripeRepo.CancelSubscription(stripe.CancelParams{
		FtcID:  ftcID,
		SubID:  subsID,
		Cancel: false,
	})

	if err != nil {
		sugar.Error(err)
		err := handleErrResp(w, err)
		if err == nil {
			return
		}

		_ = render.New(w).DBError(err)
		return
	}

	// Remember uuid to stripe subscription mapping;
	// Backup previous membership.
	if result.Modified {
		go func() {
			router.handleSubsResult(result)
		}()
	}

	_ = render.New(w).OK(result)
}
