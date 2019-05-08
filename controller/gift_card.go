package controller

import (
	"github.com/FTChinese/go-rest/view"
	"gitlab.com/ftchinese/subscription-api/model"
	"gitlab.com/ftchinese/subscription-api/paywall"
	"gitlab.com/ftchinese/subscription-api/util"
	"net/http"
)

type GiftCardRouter struct {
	model model.Env
}

// NewGiftCardRouter create a new instance of GiftCardRouter.
func NewGiftCardRouter(m model.Env) GiftCardRouter {
	return GiftCardRouter{
		model: m,
	}
}

// Redeem creates a new membership based on a gift card code.
//
//	PUT /gift-card/redeem
//
// Input {code: string}
func (router GiftCardRouter) Redeem(w http.ResponseWriter, req *http.Request)  {
	ftcID, unionID := GetUserOrUnionID(req.Header)

	code, err := util.GetJSONString(req.Body, "code")
	if err != nil {
		view.Render(w, view.NewBadRequest(err.Error()))
		return
	}
	if code != "" {
		r := view.NewReason()
		r.Field = "code"
		r.Code = view.CodeMissingField
		view.Render(w, view.NewUnprocessable(r))
		return
	}

	// Find the gift card by the card's code
	card, err := router.model.FindGiftCard(code)
	if err != nil {
		view.Render(w, view.NewDBFailure(err))
		return
	}

	// Crate a new membership from user's ftc id or
	// union id. Other fields are not set yet.
	member, err := paywall.NewMember(ftcID, unionID)
	if err != nil {
		view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	// Now the membership is updated
	member, err = member.FromGiftCard(card)
	if err != nil {
		view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	err = router.model.RedeemGiftCard(card, member)
	if err != nil {
		view.Render(w, view.NewDBFailure(err))
		return
	}

	view.Render(w, view.NewNoContent())
}