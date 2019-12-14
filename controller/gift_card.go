package controller

import (
	"github.com/FTChinese/go-rest/view"
	"github.com/jmoiron/sqlx"
	"gitlab.com/ftchinese/subscription-api/models/subscription"
	"gitlab.com/ftchinese/subscription-api/models/util"
	"gitlab.com/ftchinese/subscription-api/repository/giftrepo"
	"net/http"
)

type GiftCardRouter struct {
	env giftrepo.GiftEnv
}

// NewGiftCardRouter create a new instance of GiftCardRouter.
func NewGiftCardRouter(db *sqlx.DB, config util.BuildConfig) GiftCardRouter {
	return GiftCardRouter{
		env: giftrepo.NewGiftEnv(db, config),
	}
}

// Redeem creates a new membership based on a gift card code.
//
//	PUT /gift-card/redeem
//
// Input {code: string}
func (router GiftCardRouter) Redeem(w http.ResponseWriter, req *http.Request) {

	userID, err := GetUserID(req.Header)
	if err != nil {
		_ = view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	code, err := util.GetJSONString(req.Body, "code")
	if err != nil {
		_ = view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	// message:
	// error.field: code
	// error.code: "missing_field"
	if code == "" {
		r := view.NewReason()
		r.Field = "redeem_code"
		r.Code = view.CodeMissingField
		_ = view.Render(w, view.NewUnprocessable(r))
		return
	}

	// Find the gift card by the card's code
	// 404 Not Found
	card, err := router.env.FindGiftCard(code)
	if err != nil {
		_ = view.Render(w, view.NewDBFailure(err))
		return
	}

	// Update membership from based on gift card info.
	member, err := subscription.NewMember(userID).FromGiftCard(card)
	if err != nil {
		_ = view.Render(w, view.NewBadRequest(err.Error()))
		return
	}

	// Flag the card as already used and insert a member.
	// DB throws error if this user id already exists.
	err = router.env.RedeemGiftCard(card, member)

	if err != nil {
		// error.field: "member"
		// error.code: "already_exists"
		if util.IsAlreadyExists(err) {
			r := view.NewReason()
			r.Field = "member"
			r.Code = view.CodeAlreadyExists
			_ = view.Render(w, view.NewUnprocessable(r))
			return
		}

		_ = view.Render(w, view.NewDBFailure(err))
		return
	}

	_ = view.Render(w, view.NewNoContent())
}
