package controller

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/postoffice"
	"github.com/FTChinese/go-rest/view"
	"github.com/guregu/null"
	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/ephemeralkey"
	"gitlab.com/ftchinese/subscription-api/model"
	"gitlab.com/ftchinese/subscription-api/paywall"
	"gitlab.com/ftchinese/subscription-api/stripepay"
	"gitlab.com/ftchinese/subscription-api/util"
	"net/http"
)

type StripeRouter struct {
	PayRouter
}

func NewStripeRouter(m model.Env, p postoffice.Postman, sandbox bool) StripeRouter {
	r := StripeRouter{}

	r.sandbox = sandbox
	r.model = m
	r.postman = p

	return r
}

func stripeBadRequest(err error) view.Response {
	if stripeErr, ok := err.(*stripe.Error); ok {
		resp := view.NewResponse()
		resp.StatusCode = stripeErr.HTTPStatusCode
		resp.Body = stripeErr
		return resp
	}

	return view.NewBadRequest(err.Error())
}

func stripeDBFailure(err error) view.Response {
	if stripeErr, ok := err.(*stripe.Error); ok {
		resp := view.NewResponse()
		resp.StatusCode = stripeErr.HTTPStatusCode
		resp.Body = stripeErr
		return resp
	}

	return view.NewDBFailure(err)
}

// PlaceOrder creates an order for stripe payment.
func (router StripeRouter) PlaceOrder(w http.ResponseWriter, req *http.Request) {

	logger := logrus.WithField("trace", "StripeRouter.PlaceOrder")

	ftcID := req.Header.Get(ftcIDKey)

	// Try to find a plan based on the tier and cycle.
	plan, err := router.findPlan(req)
	// If pricing plan is not found.
	if err != nil {
		logrus.Error(err)
		view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	// Save this subscription order.
	clientApp := util.NewClientApp(req)
	userID, _ := paywall.NewUserID(null.StringFrom(ftcID), null.String{})

	order, err := router.model.CreateOrder(
		userID,
		plan,
		enum.PayMethodStripe,
		clientApp,
		null.String{})

	if err != nil {
		logger.Error(err)
		router.handleOrderErr(w, err)
		return
	}

	view.Render(w, view.NewResponse().SetBody(order))
}

// GetCustomerID send client stripe's customer id.
func (router StripeRouter) GetCustomerID(w http.ResponseWriter, req *http.Request) {
	ftcID := req.Header.Get(ftcIDKey)

	stripeID, err := router.model.CreateStripeCustomer(ftcID)

	if err != nil {
		view.Render(w, stripeDBFailure(err))
		return
	}

	view.Render(w, view.NewResponse().SetBody(map[string]string{
		"id": stripeID,
	}))
}

// ListCustomerCards lists the cards a customer owns.
func (router StripeRouter) ListCustomerCards(w http.ResponseWriter, req *http.Request) {
	id, err := GetURLParam(req, "id").ToString()
	if err != nil {
		view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	logrus.WithField("trace", "StripeRouter.ListCustomerCard").Infof("Customer id: %s", id)

	cards, err := stripepay.GetCustomerCards(id)
	if err != nil {
		view.Render(w, stripeBadRequest(err))
		return
	}

	view.Render(w, view.NewResponse().SetBody(cards))
}

// AddCard adds a bank card to a customer.
func (router StripeRouter) AddCard(w http.ResponseWriter, req *http.Request) {
	id, err := GetURLParam(req, "id").ToString()
	if err != nil {
		view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	token, err := util.GetJSONString(req.Body, "token")
	if err != nil {
		view.Render(w, view.NewBadRequest(err.Error()))
		return
	}
	if token == "" {
		reason := view.NewReason()
		reason.Field = "token"
		reason.Code = view.CodeMissingField
		view.Render(w, view.NewUnprocessable(reason))
		return
	}

	c, err := stripepay.AddCard(id, token)
	if err != nil {
		view.Render(w, stripeBadRequest(err))
		return
	}

	view.Render(w, view.NewResponse().SetBody(c))
}

// IssueKey creates an ephemeral key.
//
// GET /stripe/customers/:id/key?api_version=<version>
func (router StripeRouter) IssueKey(w http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	id, err := GetURLParam(req, "id").ToString()
	if err != nil {
		view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	stripeVersion := req.Form.Get("api_version")
	if stripeVersion == "" {
		view.Render(w, view.NewBadRequest("Stripe-Version not found"))
		return
	}

	params := &stripe.EphemeralKeyParams{
		Customer:      stripe.String(id),
		StripeVersion: stripe.String(stripeVersion),
	}
	key, err := ephemeralkey.New(params)
	if err != nil {
		view.Render(w, stripeBadRequest(err))
		return
	}

	w.Write(key.RawJSON)
}

// CreatePayIntent for stripe and returns the client secret.
// Client should already created an subscription order before hitting this endpoint.
// Client send the order id here and server uses the id to retrieve the order details, like price, to ask Stripe to create
// a payment intent.
// See: https://stripe.com/docs/payments/payment-intents/android
//
// Input: {orderId: "FTxxxxxx"}
//
// Output: {secret: "xxxxx"}
func (router StripeRouter) CreatePayIntent(w http.ResponseWriter, req *http.Request) {
	ftcID := req.Header.Get(ftcIDKey)
	orderID, err := util.GetJSONString(req.Body, "orderId")
	if err != nil {
		view.Render(w, view.NewBadRequest(err.Error()))
		return
	}
	if orderID == "" {
		reason := view.NewReason()
		reason.Field = "orderId"
		reason.Code = view.CodeMissingField
		reason.SetMessage("Order id must be provided")
		view.Render(w, view.NewUnprocessable(reason))
		return
	}

	ftcUser, err := router.model.FindFtcUser(ftcID)
	if err != nil {
		view.Render(w, view.NewDBFailure(err))
		return
	}
	if ftcUser.StripeID.IsZero() {
		view.Render(w, view.NewNotFound())
		return
	}

	billing, err := router.model.FindOrderBilling(orderID, ftcID)
	if err != nil {
		view.Render(w, view.NewDBFailure(err))
		return
	}
	if billing.IsConfirmed {
		view.Render(w, view.NewForbidden("The order provided is already confirmed."))
		return
	}

	intent, err := stripepay.CreatePaymentIntent(billing.PriceInCent(), ftcUser.StripeID.String)
	if err != nil {
		view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	view.Render(w, view.NewResponse().SetBody(map[string]string{
		"secret": intent.ClientSecret,
	}))
}
