package controller

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/stripe"
	"github.com/guregu/null"
	"net/http"
)

// CreateCheckout creates the checkout session for payment via web.
// See https://stripe.com/docs/api/checkout/sessions/create and
// https://stripe.com/docs/billing/subscriptions/checkout#create-session
// Request body:
// tier: "standard | premium"
// cycle: "month | year"
func (router StripeRouter) CreateCheckoutSession(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	ftcID := req.Header.Get(userIDKey)

	var input stripe.CheckoutInput
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

	account, err := router.accountRepo.BaseAccountByUUID(ftcID)
	if err != nil {
		_ = render.New(w).DBError(err)
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

	sess, err := router.client.NewCheckoutSession(stripe.CheckoutParams{
		Account: account,
		Plan:    sp,
		Input:   input,
	})
	if err != nil {
		err := handleErrResp(w, err)
		if err == nil {
			return
		}

		_ = render.New(w).DBError(err)
		return
	}

	if account.StripeID.IsZero() {
		account.StripeID = null.StringFrom(sess.Customer.ID)
	}

	err = router.stripeRepo.SetCustomer(account)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(stripe.NewCheckoutSession(sess))
}
