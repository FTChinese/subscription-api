package controller

import (
	"errors"
	"github.com/FTChinese/go-rest"
	"github.com/go-chi/chi"
	"github.com/guregu/null"
	log "github.com/sirupsen/logrus"
	"gitlab.com/ftchinese/subscription-api/paywall"
	"net/http"
)

const (
	wxOAuthCallback = "http://next.ftchinese.com/user/login/wechat/callback?"
)

var (
	errRenewalForbidden    = errors.New("exceed maximum allowed membership duration")
	errDowngradeNotAllowed = errors.New("membership downgrading is not allowed")
)

var logger = log.WithField("project", "subscription-api").
	WithField("package", "controller")

// GetURLParam gets a url parameter.
func GetURLParam(req *http.Request, key string) gorest.Param {
	v := chi.URLParam(req, key)

	return gorest.NewParam(key, v)
}

// GetUserOrUnionID is a convenient function to extract
// user's ftc id or wechat union id from request header.
func GetUserOrUnionID(h http.Header) (null.String, null.String) {
	uID := h.Get(userIDKey)
	wID := h.Get(unionIDKey)

	ftcID := null.NewString(uID, uID != "")
	unionID := null.NewString(wID, wID != "")

	return ftcID, unionID
}

// GetUser extract ftc uuid or union id from request header.
func GetUser(h http.Header) (paywall.User, error) {
	uID := h.Get(userIDKey)
	wID := h.Get(unionIDKey)

	ftcID := null.NewString(uID, uID != "")
	unionID := null.NewString(wID, wID != "")

	return paywall.NewUser(ftcID, unionID)
}
