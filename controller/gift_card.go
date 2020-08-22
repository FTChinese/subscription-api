package controller

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/FTChinese/subscription-api/pkg/validator"
	"github.com/FTChinese/subscription-api/repository/giftrepo"
	"github.com/jmoiron/sqlx"
	"net/http"
)

type GiftCardRouter struct {
	env giftrepo.GiftEnv
}

// NewGiftCardRouter create a new instance of GiftCardRouter.
func NewGiftCardRouter(db *sqlx.DB, config config.BuildConfig) GiftCardRouter {
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

	readerIDs := getReaderIDs(req.Header)

	code, err := GetJSONString(req.Body, "code")
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	// message:
	// error.field: code
	// error.code: "missing_field"
	if ve := validator.New("code").Required().Validate(code); ve != nil {
		_ = render.New(w).Unprocessable(ve)
		return
	}

	// Find the gift card by the card's code
	// 404 Not Found
	card, err := router.env.FindGiftCard(code)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	// Update membership from based on gift card info.
	member, err := subs.NewMember(readerIDs).FromGiftCard(card)
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	// Flag the card as already used and insert a member.
	// DB throws error if this user id already exists.
	err = router.env.RedeemGiftCard(card, member)

	if err != nil {
		// error.field: "member"
		// error.code: "already_exists"
		if db.IsAlreadyExists(err) {
			_ = render.New(w).Unprocessable(&render.ValidationError{
				Message: "duplicate entry",
				Field:   "member",
				Code:    render.CodeAlreadyExists,
			})
			return
		}

		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).NoContent()
}
