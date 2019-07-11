package controller

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/view"
	"github.com/sirupsen/logrus"
	"gitlab.com/ftchinese/subscription-api/util"
	"gitlab.com/ftchinese/subscription-api/wechat"
	"net/http"
)

type wxpayBody struct {
	PlanID string `json:"planId"`
	OpenID string `json:"openId"`
}

func (w wxpayBody) validate(tradeType wechat.TradeType) *view.Reason {
	if w.PlanID == "" {
		r := view.NewReason()
		r.Field = "planId"
		r.Code = view.CodeMissingField
		r.SetMessage("Please select a plan to subscribe")
		return r
	}

	if tradeType == wechat.TradeTypeJSAPI && w.OpenID == "" {
		r := view.NewReason()
		r.Field = "openId"
		r.Code = view.CodeMissingField
		r.SetMessage("You must provide open id to use wechat js api")

		return r
	}
	return nil
}

func (router WxPayRouter) NewSub(tradeType wechat.TradeType) http.HandlerFunc {
	logger := logrus.WithFields(logrus.Fields{
		"trace": "WxPayRouter.NewSub",
		"type":  tradeType.String(),
	})

	return func(w http.ResponseWriter, req *http.Request) {
		logger.Info("Start creating a new order with wxpay")

		var input wxpayBody
		if err := gorest.ParseJSON(req.Body, &input); err != nil {
			view.Render(w, view.NewBadRequest(err.Error()))
			return
		}

		if r := input.validate(tradeType); r != nil {
			view.Render(w, view.NewUnprocessable(r))
			return
		}

		payClient, err := router.selectClient(tradeType)
		if err != nil {
			logger.Error(err)
			view.Render(w, view.NewInternalError(err.Error()))
			return
		}

		userID, _ := GetUser(req.Header)

		plan, err := router.model.GetCurrentPricing().GetPlanByID(input.PlanID)
		if err != nil {
			logger.Error(err)
			view.Render(w, view.NewBadRequest(err.Error()))
			return
		}

		clientApp := util.NewClientApp(req)

	}
}
