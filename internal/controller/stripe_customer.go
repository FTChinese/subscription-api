package controller

import (
	"github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/stripe"
	"net/http"
)

// CreateCustomer creates stripe customer if not present.
// POST /stripe/customers
// Response: reader.FtcAccount
func (router StripeRouter) CreateCustomer(w http.ResponseWriter, req *http.Request) {
	ftcID := req.Header.Get(userIDKey)

	cusAccount, err := router.stripeRepo.CreateCustomer(ftcID)

	if err != nil {
		err := handleErrResp(w, err)
		if err == nil {
			return
		}

		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(cusAccount.Customer)
}

// CreateCustomerLegacy create a stripe customer.
// Deprecated.
func (router StripeRouter) CreateCustomerLegacy(w http.ResponseWriter, req *http.Request) {
	ftcID := req.Header.Get(userIDKey)

	cusAccount, err := router.stripeRepo.CreateCustomer(ftcID)

	if err != nil {
		err := handleErrResp(w, err)
		if err == nil {
			return
		}

		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(cusAccount.FtcAccount)
}

func (router StripeRouter) GetCustomer(w http.ResponseWriter, req *http.Request) {
	ftcID := req.Header.Get(userIDKey)
	cusID, err := getURLParam(req, "id").ToString()
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	account, err := router.readerRepo.AccountByFtcID(ftcID)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}
	if account.StripeID.IsZero() {
		_ = render.New(w).NotFound("Not a stripe customer")
		return
	}
	if account.StripeID.String != cusID {
		_ = render.New(w).NotFound("")
	}

	cus, err := router.client.RetrieveCustomer(account.StripeID.String)
	if err != nil {
		err := handleErrResp(w, err)
		if err == nil {
			return
		}

		_ = render.NewInternalError(err.Error())
		return
	}

	_ = render.New(w).OK(stripe.NewCustomer(account, cus))
}

func (router StripeRouter) ChangeDefaultPaymentMethod(w http.ResponseWriter, req *http.Request) {
	ftcID := req.Header.Get(userIDKey)
	cusID, err := getURLParam(req, "id").ToString()
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	var pm stripe.PaymentInput
	if err := gorest.ParseJSON(req.Body, &pm); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}
	pm.CustomerID = cusID

	if ve := pm.Validate(); ve != nil {
		_ = render.New(w).Unprocessable(ve)
		return
	}

	account, err := router.readerRepo.AccountByFtcID(ftcID)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}
	if account.StripeID.IsZero() {
		_ = render.New(w).NotFound("Not a stripe customer")
		return
	}
	if account.StripeID.String != cusID {
		_ = render.New(w).NotFound("")
	}

	cus, err := router.client.SetDefaultPaymentMethod(pm)
	if err != nil {
		err = handleErrResp(w, err)
		if err != nil {
			_ = render.New(w).BadRequest(err.Error())
		}
		return
	}

	_ = render.New(w).OK(stripe.NewCustomer(account, cus))
}

// GetDefaultPaymentMethod retrieves a user's invoice default payment method.
// GET /stripe/customers/{id}/default_payment_method
// Deprecated
func (router StripeRouter) GetDefaultPaymentMethod(w http.ResponseWriter, req *http.Request) {
	// Get stripe customer id from url.
	id, err := getURLParam(req, "id").ToString()
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	cus, err := router.client.RetrieveCustomer(id)
	if err != nil {
		err = handleErrResp(w, err)
		if err != nil {
			_ = render.New(w).DBError(err)
		}
		return
	}

	if cus.InvoiceSettings == nil || cus.InvoiceSettings.DefaultPaymentMethod == nil {
		_ = render.New(w).NotFound("Default payment method not set yet")
		return
	}

	_ = render.New(w).OK(cus.InvoiceSettings.DefaultPaymentMethod)
}

// SetDefaultPaymentMethod sets stripe customer's invoice_settings.default_payment_method.
// POST /stripe/customers/{id}/default_payment_method
//
// Input: {defaultPaymentMethod: string}. The id of the default payment method.
// Deprecated
func (router StripeRouter) SetDefaultPaymentMethod(w http.ResponseWriter, req *http.Request) {
	// Get stripe customer id from url.
	cusID, err := getURLParam(req, "id").ToString()
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	var pm stripe.PaymentInput
	if err := gorest.ParseJSON(req.Body, &pm); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}
	pm.CustomerID = cusID

	if ve := pm.Validate(); ve != nil {
		_ = render.New(w).Unprocessable(ve)
		return
	}

	cus, err := router.client.SetDefaultPaymentMethod(pm)
	if err != nil {
		err = handleErrResp(w, err)
		if err != nil {
			_ = render.New(w).BadRequest(err.Error())
		}
		return
	}

	_ = render.New(w).OK(cus)
}
