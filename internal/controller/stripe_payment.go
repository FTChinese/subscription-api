package controller

import (
	"github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	ftcStripe "github.com/FTChinese/subscription-api/pkg/stripe"
	"net/http"
)

// GetDefaultPaymentMethod retrieves a user's invoice default payment method.
// GET /stripe/customers/{id}/default_payment_method
func (router StripeRouter) GetDefaultPaymentMethod(w http.ResponseWriter, req *http.Request) {
	// Get stripe customer id from url.
	id, err := getURLParam(req, "id").ToString()
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	cus, err := router.client.RetrieveCustomer(id)
	if err != nil {
		err = forwardStripeErr(w, err)
		if err != nil {
			_ = render.New(w).DBError(err)
		}
		return
	}

	if cus.InvoiceSettings == nil {
		_ = render.New(w).NotFound()
		return
	}

	_ = render.New(w).OK(cus.InvoiceSettings.DefaultPaymentMethod)
}

// SetDefaultPaymentMethod sets stripe customer's invoice_settings.default_payment_method.
// POST /stripe/customers/{id}/default_payment_method
//
// Input: {defaultPaymentMethod: string}. The id of the default payment method.
func (router StripeRouter) SetDefaultPaymentMethod(w http.ResponseWriter, req *http.Request) {
	// Get stripe customer id from url.
	id, err := getURLParam(req, "id").ToString()
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	var pm ftcStripe.PaymentInput
	if err := gorest.ParseJSON(req.Body, &pm); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}
	pm.CustomerID = id

	if ve := pm.Validate(); ve != nil {
		_ = render.New(w).Unprocessable(ve)
		return
	}

	cus, err := router.client.SetDefaultPaymentMethod(pm)
	if err != nil {
		err = forwardStripeErr(w, err)
		if err != nil {
			_ = render.New(w).BadRequest(err.Error())
		}
		return
	}

	_ = render.New(w).OK(cus)
}
