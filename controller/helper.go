package controller

import (
	"github.com/FTChinese/go-rest"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
	"net/http"
)

const (
	alipayCallback  = "http://next.ftchinese.com/user/subscription/alipay/callback?"
	wxOAuthCallback = "http://next.ftchinese.com/user/login/callback?"
)

var logger = log.WithField("project", "subscription-api").
	WithField("package", "controller")

// GetURLParam gets a url parameter.
func GetURLParam(req *http.Request, key string) gorest.Param {
	v := chi.URLParam(req, key)

	return gorest.NewParam(key, v)
}
