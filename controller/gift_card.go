package controller

import (
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/repository/giftrepo"
	"github.com/jmoiron/sqlx"
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
