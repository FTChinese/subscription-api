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
func (routes StripeRoutes) CreateCustomer(w http.ResponseWriter, req *http.Request) {
	defer routes.logger.Sync()
	sugar := routes.logger.Sugar()

	ftcID := xhttp.GetFtcID(req.Header)

	cus, err := routes.stripeRepo.CreateCustomer(ftcID)

	if err != nil {
		sugar.Error(err)
		_ = xhttp.HandleSubsErr(w, err)
		return
	}

	if cus.IsFromStripe {
		go func() {
			err := routes.stripeRepo.UpsertCustomer(cus)
			if err != nil {
				sugar.Error(err)
			}
		}()
	}

	_ = render.New(w).OK(cus)
}

func (routes StripeRoutes) GetCustomer(w http.ResponseWriter, req *http.Request) {
	defer routes.logger.Sync()
	sugar := routes.logger.Sugar()

	ftcID := xhttp.GetFtcID(req.Header)
	refresh := xhttp.ParseQueryRefresh(req)
	cusID, err := xhttp.GetURLParam(req, "id").ToString()
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	cus, err := routes.loadCustomer(cusID, refresh)
	if err != nil {
		sugar.Error(err)
		_ = xhttp.HandleSubsErr(w, err)
		return
	}

	// If the request ftc id and the found customer mismatched.
	if cus.FtcID != ftcID {
		_ = render.New(w).NotFound("")
		return
	}

	_ = render.New(w).OK(cus)
}

func (routes StripeRoutes) loadCustomer(id string, refresh bool) (stripe.Customer, error) {
	defer routes.logger.Sync()
	sugar := routes.logger.Sugar()

	cus, err := routes.stripeRepo.LoadOrFetchCustomer(id, refresh)
	if err != nil {
		sugar.Error(err)
		return stripe.Customer{}, err
	}

	if !cus.IsFromStripe {
		return cus, nil
	}

	ba, err := routes.readerRepo.BaseAccountByStripeID(id)
	if err != nil {
		sugar.Error(err)
		return stripe.Customer{}, err
	}

	cus = cus.WithFtcID(ba.FtcID)

	go func() {
		err := routes.stripeRepo.UpsertCustomer(cus)
		if err != nil {
			sugar.Error(err)
		}
	}()

	return cus, nil
}

// GetCusDefaultPaymentMethod load the payment method details
// which is set as a customer's default payment method.
func (routes StripeRoutes) GetCusDefaultPaymentMethod(w http.ResponseWriter, req *http.Request) {
	defer routes.logger.Sync()
	sugar := routes.logger.Sugar()

	cusID, err := xhttp.GetURLParam(req, "id").ToString()
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	refresh := xhttp.ParseQueryRefresh(req)

	// Load customer first; otherwise we do not know the
	// default payment method id.
	cus, err := routes.loadCustomer(cusID, false)
	if err != nil {
		sugar.Error(err)
		_ = xhttp.HandleSubsErr(w, err)
		return
	}

	// Default payment method does not exist.
	if cus.DefaultPaymentMethodID.IsZero() {
		_ = render.New(w).NotFound("Default payment method not set")
		return
	}

	// Fetch payment method
	pm, err := routes.loadPaymentMethod(
		cus.DefaultPaymentMethodID.String,
		refresh)
	if err != nil {
		sugar.Error(err)
		_ = xhttp.HandleSubsErr(w, err)
		return
	}

	_ = render.New(w).OK(pm)
}

func (routes StripeRoutes) UpdateCusDefaultPaymentMethod(w http.ResponseWriter, req *http.Request) {
	defer routes.logger.Sync()
	sugar := routes.logger.Sugar()

	ftcID := xhttp.GetFtcID(req.Header)
	cusID, err := xhttp.GetURLParam(req, "id").ToString()
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

	acnt, err := routes.readerRepo.BaseAccountByUUID(ftcID)
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

	rawCus, err := routes.stripeRepo.Client.SetCusDefaultPaymentMethod(
		cusID,
		params.ID)
	if err != nil {
		err = xhttp.HandleSubsErr(w, err)
		return
	}

	cus := stripe.NewCustomer(acnt.FtcID, rawCus)

	// Update payment method in db.
	// We do not update customer here since the webhook
	// will handle it.
	go func() {
		// Save updated customer
		err := routes.stripeRepo.UpsertCustomer(cus)
		if err != nil {
			sugar.Error(err)
		}

		// Fetch the related payment method from Stripe
		pm, err := routes.stripeRepo.LoadOrFetchPaymentMethod(params.ID, true)
		if err != nil {
			sugar.Error(err)
			return
		}

		// Save this payment method.
		err = routes.stripeRepo.UpsertPaymentMethod(pm)
		if err != nil {
			sugar.Error(err)
		}
	}()

	_ = render.New(w).OK(cus)
}

func (routes StripeRoutes) ListCusPaymentMethods(w http.ResponseWriter, req *http.Request) {
	defer routes.logger.Sync()
	sugar := routes.logger.Sugar()

	p := gorest.GetPagination(req)
	cusID, err := xhttp.GetURLParam(req, "id").ToString()
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	list, err := routes.stripeRepo.ListCusPaymentMethods(cusID, p)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(list)
}
