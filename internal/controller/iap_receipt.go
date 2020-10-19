package controller

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/repository/iaprepo"
	"github.com/FTChinese/subscription-api/pkg/apple"
	"net/http"
)

// LoadReceipt retrieves the subscription data for
// an original transaction id, together with the
// receipt used to verify it.
func (router IAPRouter) LoadReceipt(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	id, _ := getURLParam(req, "id").ToString()

	sub, err := router.iapRepo.LoadSubs(id)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	b, err := iaprepo.LoadReceipt(sub.BaseSchema)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).NotFound()
		return
	}

	data := struct {
		apple.Subscription
		Receipt string `json:"receipt"`
	}{
		Subscription: sub,
		Receipt:      string(b),
	}

	_ = render.New(w).OK(data)
}
