package controller

import (
	"github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/view"
	"github.com/go-chi/chi"
	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go"
	"gitlab.com/ftchinese/subscription-api/models/reader"
	"gitlab.com/ftchinese/subscription-api/pkg/product"
	"net/http"
)

const (
	wxOAuthCallback = "http://users.ftchinese.com/login/wechat/callback?"
	wxAppNativeApp  = "wxacddf1c20516eb69" // Used by native app to pay and log in.
)

var logger = logrus.WithField("project", "subscription-api").WithField("package", "controller")

// GetURLParam gets a url parameter.
func GetURLParam(req *http.Request, key string) gorest.Param {
	v := chi.URLParam(req, key)

	return gorest.NewParam(key, v)
}

func GetEdition(req *http.Request) (product.Edition, error) {
	t, err := GetURLParam(req, "tier").ToString()
	if err != nil {
		return product.Edition{}, err
	}

	tier, err := enum.ParseTier(t)
	if err != nil {
		return product.Edition{}, err
	}

	c, err := GetURLParam(req, "cycle").ToString()
	if err != nil {
		return product.Edition{}, err
	}

	cycle, err := enum.ParseCycle(c)
	if err != nil {
		return product.Edition{}, err
	}

	return product.Edition{
		Tier:  tier,
		Cycle: cycle,
	}, nil
}

// getWxAppID from query parameter, and fallback to request header, then fallback to hard-coded one.
func getWxAppID(req *http.Request) string {
	appID := req.FormValue("app_id")
	if appID != "" {
		return appID
	}

	// Prior to v0.8.0 the app id is set in header.
	appID = req.Header.Get(appIDKey)
	if appID != "" {
		return appID
	}

	// For backward compatibility with Android <= 2.0.4
	return wxAppNativeApp
}

// GetUserID extract ftc uuid or union id from request header.
func GetUserID(h http.Header) (reader.MemberID, error) {
	return reader.NewMemberID(h.Get(ftcIDKey), h.Get(unionIDKey))
}

// CastStripeError tries to cast an error to stripe.Error, or nil if it is not.
func CastStripeError(err error) *stripe.Error {
	if stripeErr, ok := err.(*stripe.Error); ok {
		return stripeErr
	}

	return nil
}

func BuildStripeResponse(e *stripe.Error) view.Response {
	r := view.NewResponse()
	r.StatusCode = e.HTTPStatusCode
	r.Body = view.ClientError{
		Message: e.Msg,
		Code:    string(e.Code),
		Param:   e.Param,
		Type:    string(e.Type),
	}

	return r
}
