package controller

import (
	"github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/view"
	"github.com/go-chi/chi"
	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go"
	"gitlab.com/ftchinese/subscription-api/models/reader"
	"net/http"
)

const (
	wxOAuthCallback = "http://next.ftchinese.com/user/login/wechat/callback?"
)

var logger = logrus.WithField("project", "subscription-api").WithField("package", "controller")

// GetURLParam gets a url parameter.
func GetURLParam(req *http.Request, key string) gorest.Param {
	v := chi.URLParam(req, key)

	return gorest.NewParam(key, v)
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
