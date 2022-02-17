package api

import (
	"github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/pkg/stripe"
	"github.com/FTChinese/subscription-api/pkg/xhttp"
	"net/http"
)

// CreateCustomer creates stripe customer for a ftc user id.
// We first retrieve a user's account to see if a stripe customer id exists.
// If it does, then we return the customer data directly;
// a new customer is created and returned.
// The customer data might come from FTC's db, or fetched from Stripe API.
// We shall save a customer in the latter case so that we don't need
// to hit Stripe each time the data is requested.
// Webhook should listen for customer updated event. See WebHook method.
//
// POST /stripe/customers
// Request: empty.
// Response: stripe.Customer
func (router StripeRouter) CreateCustomer(w http.ResponseWriter, req *http.Request) {
	defer router.Logger.Sync()
	sugar := router.Logger.Sugar()

	ftcID := xhttp.GetFtcID(req.Header)

	cus, err := router.Env.CreateCustomer(ftcID)

	if err != nil {
		sugar.Error(err)
		_ = xhttp.HandleStripeErr(w, err)
		return
	}

	if cus.IsFromStripe {
		go func() {
			err := router.Env.UpsertCustomer(cus)
			if err != nil {
				sugar.Error(err)
			}
		}()
	}

	_ = render.New(w).OK(cus)
}

func (router StripeRouter) GetCustomer(w http.ResponseWriter, req *http.Request) {
	defer router.Logger.Sync()
	sugar := router.Logger.Sugar()

	ftcID := xhttp.GetFtcID(req.Header)
	cusID, err := xhttp.GetURLParam(req, "id").ToString()

	if err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	cus, err := router.getCustomer(cusID)
	if err != nil {
		sugar.Error(err)
		_ = xhttp.HandleStripeErr(w, err)
		return
	}

	// If the request ftc id and the found customer mismatched.
	if cus.FtcID != ftcID {
		_ = render.New(w).NotFound("")
		return
	}

	if cus.IsFromStripe {
		go func() {
			err := router.Env.UpsertCustomer(cus)
			if err != nil {
				sugar.Error(err)
			}
		}()
	}

	_ = render.New(w).OK(cus)
}

func (router StripeRouter) getCustomer(id string) (stripe.Customer, error) {
	cus, err := router.Env.LoadOrFetchCustomer(id)
	if err == nil {
		return cus, nil
	}

	// If this customer is not found in our db, stop hitting Stripe API.
	baseAccount, err := router.ReaderRepo.BaseAccountByStripeID(id)
	if err != nil {
		return stripe.Customer{}, nil
	}

	return cus.WithFtcID(baseAccount.FtcID), nil
}

// GetCusDefaultPaymentMethod load the payment method details
// which is set as a customer's default payment method.
func (router StripeRouter) GetCusDefaultPaymentMethod(w http.ResponseWriter, req *http.Request) {
	defer router.Logger.Sync()
	sugar := router.Logger.Sugar()

	cusID, err := xhttp.GetURLParam(req, "id").ToString()
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	// Load customer first; otherwise we do not know the
	// default payment method id.
	cus, err := router.getCustomer(cusID)
	if err != nil {
		sugar.Error(err)
		_ = xhttp.HandleStripeErr(w, err)
		return
	}

	// Save to our db.
	if cus.IsFromStripe {
		go func() {
			err := router.Env.UpsertCustomer(cus)
			if err != nil {
				sugar.Error(err)
			}
		}()
	}

	// Default payment method does not exist.
	if cus.DefaultPaymentMethodID.IsZero() {
		_ = render.New(w).NotFound("Default payment method not set")
		return
	}

	// Fetch payment method
	pm, err := router.Env.LoadOrFetchPaymentMethod(cus.DefaultPaymentMethodID.String)
	if err != nil {
		sugar.Error(err)
		_ = xhttp.HandleStripeErr(w, err)
		return
	}

	// Save it if not save in our db yet.
	if pm.IsFromStripe {
		go func() {
			err := router.Env.UpsertPaymentMethod(pm)
			if err != nil {
				sugar.Error(err)
			}
		}()
	}

	_ = render.New(w).OK(pm)
}

func (router StripeRouter) UpdateCusDefaultPaymentMethod(w http.ResponseWriter, req *http.Request) {
	defer router.Logger.Sync()
	sugar := router.Logger.Sugar()

	ftcID := xhttp.GetFtcID(req.Header)
	cusID, err := xhttp.GetURLParam(req, "id").ToString()
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	var params stripe.DefaultPaymentMethodParams
	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	if ve := params.Validate(); ve != nil {
		_ = render.New(w).Unprocessable(ve)
		return
	}

	acnt, err := router.ReaderRepo.BaseAccountByUUID(ftcID)
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

	rawCus, err := router.Env.Client.SetCusDefaultPaymentMethod(
		cusID,
		params.DefaultMethod)
	if err != nil {
		err = xhttp.HandleStripeErr(w, err)
		return
	}

	cus := stripe.NewCustomer(acnt.FtcID, rawCus)

	// Update payment method in db.
	// We do not update customer here since the webhook
	// will handle it.
	go func() {
		err := router.Env.UpsertCustomer(cus)
		if err != nil {
			sugar.Error(err)
		}

		_, err = router.Env.FetchAndSavePaymentMethod(params.DefaultMethod)
		if err != nil {
			sugar.Error(err)
		}
	}()

	_ = render.New(w).OK(cus)
}

func (router StripeRouter) ListCusPaymentMethods(w http.ResponseWriter, req *http.Request) {
	defer router.Logger.Sync()
	sugar := router.Logger.Sugar()

	p := gorest.GetPagination(req)
	cusID, err := xhttp.GetURLParam(req, "id").ToString()
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	list, err := router.Env.ListCusPaymentMethods(cusID, p)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(list)
}
