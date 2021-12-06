package controller

import (
	"github.com/FTChinese/subscription-api/internal/repository/giftrepo"
	"github.com/jmoiron/sqlx"
)

type GiftCardRouter struct {
	env giftrepo.Env
}

// NewGiftCardRouter create a new instance of GiftCardRouter.
func NewGiftCardRouter(db *sqlx.DB) GiftCardRouter {
	return GiftCardRouter{
		env: giftrepo.New(db),
	}
}
