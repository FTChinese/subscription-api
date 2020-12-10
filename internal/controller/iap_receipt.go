package controller

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/apple"
	"net/http"
)

// LoadReceipt retrieves the subscription data for
// an original transaction id, together with the
// receipt used to verify it.
func (router IAPRouter) LoadReceipt(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	origTxID, err := getURLParam(req, "id").ToString()
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	fsOnly := req.FormValue("fs") == "true"

	sub, err := router.iapRepo.LoadSubs(origTxID)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	receipt, err := router.iapRepo.LoadReceipt(sub.BaseSchema, fsOnly)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).NotFound("Apple subscription not found")
		return
	}

	data := struct {
		apple.Subscription
		Receipt string `json:"receipt"`
	}{
		Subscription: sub,
		Receipt:      receipt,
	}

	_ = render.New(w).OK(data)
}
