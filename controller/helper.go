package controller

import (
	"github.com/FTChinese/go-rest"
	"github.com/go-chi/chi"
	"github.com/guregu/null"
	"gitlab.com/ftchinese/subscription-api/paywall"
	"net/http"
)

const (
	wxOAuthCallback = "http://next.ftchinese.com/user/login/wechat/callback?"
)

// GetURLParam gets a url parameter.
func GetURLParam(req *http.Request, key string) gorest.Param {
	v := chi.URLParam(req, key)

	return gorest.NewParam(key, v)
}

// GetUser extract ftc uuid or union id from request header.
func GetUser(h http.Header) (paywall.User, error) {
	uID := h.Get(ftcIDKey)
	wID := h.Get(unionIDKey)

	ftcID := null.NewString(uID, uID != "")
	unionID := null.NewString(wID, wID != "")

	return paywall.NewUser(ftcID, unionID)
}
