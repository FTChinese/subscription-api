package api

import (
	"github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/pkg/stripe"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/xhttp"
	"net/http"
)

// CreateSubs create a stripe subscription.
// Input:
// - priceId: string - The stripe price id to subscribe
// - introductoryPriceId: string - A one-time stripe price id to create an extra invoice
// - coupon?: string;
// - defaultPaymentMethod?: string;
// - idempotency?: string;
//
// PITFALLS:
// If you create a plan in CNY, and a customer is subscribed to
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
func (routes StripeRoutes) CreateSubs(w http.ResponseWriter, req *http.Request) {
	defer routes.logger.Sync()
	sugar := routes.logger.Sugar()

	// Get FTC id. Its presence is already checked by middleware.
	ftcID := xhttp.GetFtcID(req.Header)
	var params stripe.SubsParams
	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}
	// Validate params data.
	if err := params.Validate(); err != nil {
		sugar.Error(err)
		_ = render.New(w).Unprocessable(err)
		return
	}

	acnt, err := routes.readerRepo.BaseAccountByUUID(ftcID)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}
	// If this user is not a stripe customer yet.
	if acnt.StripeID.IsZero() {
		_ = render.New(w).NotFound("Must be a stripe customer prior to subscription")
		return
	}

	item, err := routes.findCartItem(params)
	if err != nil {
		sugar.Error(err)
		_ = xhttp.HandleSubsErr(w, err)
		return
	}

	if ve := item.Validate(); ve != nil {
		sugar.Error(err)
		_ = render.New(w).Unprocessable(ve)
		return
	}

	cart := reader.NewShoppingCart(acnt).
		WithStripeItem(item)

	// Create stripe subscription.
	cart, result, err := routes.stripeRepo.CreateSubscription(cart, params)

	// Shopping session should be saved regardless of success or failure
	session := stripe.NewShoppingSession(cart, params)

	if err != nil {
		sugar.Error(err)
		_ = xhttp.HandleSubsErr(w, reader.ConvertIntentError(err))

		go func() {
			routes.saveShoppingSession(session)
		}()

		return
	}

	// Save ftc id to stripe subscription id mapping.
	// Backup previous membership if exists.
	go func() {
		routes.handleSubsResult(result)
		routes.saveShoppingSession(session.WithSubs(result.Subs))
	}()

	_ = render.New(w).OK(result)
}

// LoadSubs from ftc db only. If you want to refresh the subscription,
// use the refresh endpoint.
// Refreshing is not supported here as other endpoint with the
// `refresh=true` query parameter since it involves syncing
// membership.
func (routes StripeRoutes) LoadSubs(w http.ResponseWriter, req *http.Request) {

	defer routes.logger.Sync()
	sugar := routes.logger.Sugar()

	ftcID := xhttp.GetFtcID(req.Header)

	subsID, err := xhttp.GetURLParam(req, "id").ToString()
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	// We cannot use LoadOrFetchSubs here since if the data
	// is fetched from Stripe API, we cannot ensure this
	// ftc id definitely belong to this subscription.
	subs, err := routes.stripeRepo.RetrieveSubs(subsID)
	if err != nil {
		sugar.Error(err)
		_ = xhttp.HandleSubsErr(w, err)
		return
	}

	if subs.FtcUserID.String != ftcID {
		_ = render.New(w).NotFound("Subscription does not exist for this user")
		return
	}

	_ = render.New(w).OK(subs)
}

// UpdateSubs updates a stripe subscription:
// User could switch cycle of the same tier, or upgrading to a higher tier.
// Input:
// * priceId: string - The price to change to.
// * coupon?: "",
// * defaultPaymentMethod?: ""
// * idempotency?: string
//
// Error response:
// 404 if membership if not found.
// NOTE: when updating a stripe subscription, the return payload
// `items` field contains more than one items:
// one is standard and another if premium.
// So we cannot rely on this field to find FTC plan.
func (routes StripeRoutes) UpdateSubs(w http.ResponseWriter, req *http.Request) {
	defer routes.logger.Sync()
	sugar := routes.logger.Sugar()

	// Get FTC id. Its presence is already checked by middleware.
	ftcID := xhttp.GetFtcID(req.Header)
	var params stripe.SubsParams
	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}
	// Validating if user is updating to the same price is postponed to
	// building checkout intent.
	if ve := params.Validate(); ve != nil {
		_ = render.New(w).Unprocessable(ve)
		return
	}

	account, err := routes.readerRepo.BaseAccountByUUID(ftcID)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}
	// If this user is not a stripe customer yet.
	if account.StripeID.IsZero() {
		_ = render.New(w).NotFound("Stripe customer not found")
		return
	}

	item, err := routes.findCartItem(params)
	if err != nil {
		sugar.Error(err)
		_ = xhttp.HandleSubsErr(w, err)
		return
	}

	cart := reader.NewShoppingCart(account).WithStripeItem(item)
	cart, result, err := routes.stripeRepo.UpdateSubscription(
		cart,
		params,
	)

	session := stripe.NewShoppingSession(cart, params)
	if err != nil {
		sugar.Error(err)
		_ = xhttp.HandleSubsErr(w, reader.ConvertIntentError(err))
		go func() {
			routes.saveShoppingSession(session)
		}()
		return
	}

	// Remember uuid to stripe subscription mapping;
	// Backup previous membership.
	go func() {
		routes.handleSubsResult(result)
		routes.saveShoppingSession(session.WithSubs(result.Subs))
	}()

	if result.Subs.PaymentIntent.IsZero() {
		_ = render.New(w).BadRequest("PaymentIntent not expanded")
		return
	}

	_ = render.New(w).OK(result)
}

// RefreshSubs get the latest data of a subscription if user manually requested it.
func (routes StripeRoutes) RefreshSubs(w http.ResponseWriter, req *http.Request) {
	defer routes.logger.Sync()
	sugar := routes.logger.Sugar()

	// Get the subscription id from url
	subsID, err := xhttp.GetURLParam(req, "id").ToString()
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	// Use Stripe SDK to retrieve data.
	// The latest invoice field is expanded.
	ss, err := routes.stripeRepo.Client.FetchSubs(subsID, true)
	if err != nil {
		sugar.Error(err)
		err = xhttp.HandleSubsErr(w, err)
		return
	}

	// Use Stripe customer id to find user account.
	ba, err := routes.readerRepo.BaseAccountByStripeID(ss.Customer.ID)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).NotFound("Stripe customer not found")
		return
	}

	result, err := routes.stripeRepo.RefreshSubscription(ss, ba)

	if err != nil {
		sugar.Error(err)
		_ = xhttp.HandleSubsErr(w, err)
		return
	}

	// Only update subs and snapshot if actually modified.
	go func() {
		routes.handleSubsResult(result)
	}()

	_ = render.New(w).OK(result)
}

// CancelSubs cancels a stripe subscription at period end.
// See https://stripe.com/docs/billing/subscriptions/cancel
func (routes StripeRoutes) CancelSubs(w http.ResponseWriter, req *http.Request) {
	defer routes.logger.Sync()
	sugar := routes.logger.Sugar()

	ftcID := xhttp.GetFtcID(req.Header)

	subsID, err := xhttp.GetURLParam(req, "id").ToString()
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	result, err := routes.stripeRepo.CancelSubscription(stripe.CancelParams{
		FtcID:  ftcID,
		SubID:  subsID,
		Cancel: true,
	})

	if err != nil {
		sugar.Error(err)
		_ = xhttp.HandleSubsErr(w, err)
		return
	}

	// Remember uuid to stripe subscription mapping;
	// Backup previous membership.
	go func() {
		routes.handleSubsResult(result)
	}()

	_ = render.New(w).OK(result)
}

// ReactivateSubscription undo subscription cancellation before period ends.
func (routes StripeRoutes) ReactivateSubscription(w http.ResponseWriter, req *http.Request) {
	defer routes.logger.Sync()
	sugar := routes.logger.Sugar()

	ftcID := xhttp.GetFtcID(req.Header)

	subsID, err := xhttp.GetURLParam(req, "id").ToString()
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	result, err := routes.stripeRepo.CancelSubscription(stripe.CancelParams{
		FtcID:  ftcID,
		SubID:  subsID,
		Cancel: false,
	})

	if err != nil {
		sugar.Error(err)
		_ = xhttp.HandleSubsErr(w, err)
		return
	}

	// Remember uuid to stripe subscription mapping;
	// Backup previous membership.
	if result.Modified {
		go func() {
			routes.handleSubsResult(result)
		}()
	}

	_ = render.New(w).OK(result)
}

func (routes StripeRoutes) GetSubsDefaultPaymentMethod(w http.ResponseWriter, req *http.Request) {
	defer routes.logger.Sync()
	sugar := routes.logger.Sugar()

	refresh := xhttp.ParseQueryRefresh(req)

	subsID, err := xhttp.GetURLParam(req, "id").ToString()
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	subs, err := routes.stripeRepo.LoadOrFetchSubs(subsID, false)
	if err != nil {
		sugar.Error(err)
		_ = xhttp.HandleSubsErr(w, err)
		return
	}

	if subs.DefaultPaymentMethodID.IsZero() {
		_ = render.New(w).NotFound("Default payment method not set")
		return
	}

	pm, err := routes.loadPaymentMethod(subs.DefaultPaymentMethodID.String, refresh)
	if err != nil {
		sugar.Error(err)
		_ = xhttp.HandleSubsErr(w, err)
		return
	}

	_ = render.New(w).OK(pm)
}

func (routes StripeRoutes) UpdateSubsDefaultPayMethod(w http.ResponseWriter, req *http.Request) {
	defer routes.logger.Sync()
	sugar := routes.logger.Sugar()

	subsID, err := xhttp.GetURLParam(req, "id").ToString()
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	var params stripe.DefaultPaymentMethodParams
	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	if ve := params.Validate(); ve != nil {
		sugar.Error(err)
		_ = render.New(w).Unprocessable(ve)
		return
	}

	subs, err := routes.stripeRepo.LoadOrFetchSubs(subsID, false)
	if err != nil {
		_ = xhttp.HandleSubsErr(w, err)
		return
	}

	// Ensure the payment method exists
	pm, err := routes.loadPaymentMethod(params.ID, false)
	if err != nil {
		_ = xhttp.HandleSubsErr(w, err)
		return
	}

	rawSubs, err := routes.stripeRepo.Client.SetSubsDefaultPaymentMethod(
		subs.ID,
		pm.ID)
	if err != nil {
		_ = xhttp.HandleSubsErr(w, err)
		return
	}

	subs = stripe.NewSubs(subs.FtcUserID.String, rawSubs)

	go func() {
		err := routes.stripeRepo.UpsertSubs(subs, false)
		if err != nil {
			sugar.Error(err)
		}
	}()

	_ = render.New(w).OK(subs)
}
