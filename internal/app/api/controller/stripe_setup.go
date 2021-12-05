package controller

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"net/http"
)

func (router StripeRouter) CreateSetupIntent(w http.ResponseWriter, req *http.Request) {
	ftcID := ids.GetFtcID(req.Header)

	acnt, err := router.ReaderRepo.BaseAccountByUUID(ftcID)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}
	if acnt.StripeID.IsZero() {
		_ = render.New(w).BadRequest("Not a Stripe customer yet")
		return
	}

	si, err := router.Client.NewSetupCheckout(acnt.StripeID.String)
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
