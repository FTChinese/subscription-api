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

func (router StripeRouter) GetCustomer(w http.ResponseWriter, req *http.Request) {
	ftcID := req.Header.Get(userIDKey)
	cusID, err := getURLParam(req, "id").ToString()
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	account, err := router.stripeRepo.BaseAccountByUUID(ftcID)
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

	acnt, err := router.stripeRepo.BaseAccountByUUID(ftcID)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}
	if acnt.StripeID.IsZero() {
		_ = render.New(w).NotFound("Not a stripe customer")
		return
	}
	if acnt.StripeID.String != cusID {
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

	_ = render.New(w).OK(stripe.NewCustomer(acnt, cus))
}
