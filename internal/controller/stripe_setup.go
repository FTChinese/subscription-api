package controller

import (
	"github.com/FTChinese/go-rest/render"
	"net/http"
)

func (router StripeRouter) CreateSetupIntent(w http.ResponseWriter, req *http.Request) {
	ftcID := req.Header.Get(userIDKey)

	account, err := router.readerRepo.AccountByFtcID(ftcID)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}
	if account.StripeID.IsZero() {
		_ = render.New(w).BadRequest("Not a Stripe customer yet")
		return
	}

	si, err := router.client.CreateSetupIntent(account.StripeID.String)
	if err != nil {
		err := handleErrResp(w, err)
		if err == nil {
			return
		}
		_ = render.NewInternalError(err.Error())
		return
	}

	_ = render.New(w).OK(map[string]string{
		"clientSecret": si.ClientSecret,
	})
}
