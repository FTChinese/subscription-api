package controller

import (
	"github.com/FTChinese/go-rest"
	"github.com/go-chi/chi"
	"github.com/gorilla/schema"
	"net/http"
)

var decoder = schema.NewDecoder()

// getURLParam gets a url parameter.
func getURLParam(req *http.Request, key string) gorest.Param {
	v := chi.URLParam(req, key)

	return gorest.NewParam(key, v)
}
