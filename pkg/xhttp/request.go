package xhttp

import (
	"github.com/FTChinese/go-rest"
	"github.com/go-chi/chi"
	"net/http"
)

// GetURLParam gets a url parameter.
func GetURLParam(req *http.Request, key string) gorest.Param {
	v := chi.URLParam(req, key)

	return gorest.NewParam(key, v)
}
