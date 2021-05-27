package controller

import (
	"github.com/FTChinese/go-rest"
	"github.com/FTChinese/subscription-api/pkg"
	"github.com/go-chi/chi"
	"github.com/gorilla/schema"
	"github.com/guregu/null"
	"net/http"
)

const (
	wxOAuthCallback = "https://users.ftchinese.com/login/wechat/callback?"
)

var decoder = schema.NewDecoder()

// getURLParam gets a url parameter.
func getURLParam(req *http.Request, key string) gorest.Param {
	v := chi.URLParam(req, key)

	return gorest.NewParam(key, v)
}

// getReaderIDs extract ftc uuid or union id from request header.
// It panic if both ftc id and union id are missing.
// However it won't happen since middlewares already ensured at least one of them should exist.
func getReaderIDs(h http.Header) pkg.UserIDs {
	ftcID := h.Get(ftcIDKey)
	unionID := h.Get(unionIDKey)

	return pkg.UserIDs{
		CompoundID: "",
		FtcID:      null.NewString(ftcID, ftcID != ""),
		UnionID:    null.NewString(unionID, unionID != ""),
	}.MustNormalize()
}
