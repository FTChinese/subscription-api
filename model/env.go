package model

import (
	"database/sql"

	"github.com/FTChinese/go-rest/enum"

	cache "github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
)

// Env wraps database connection
type Env struct {
	DB    *sql.DB
	Cache *cache.Cache
}

var logger = log.WithField("project", "subscription-api").WithField("package", "model")

const (
	keySchedule = "discountSchedule"
	keyPromo    = "promotionSchedule"
)

func normalizeMemberTier(vipType int64) enum.Tier {
	switch vipType {

	case 10:
		return enum.TierStandard

	case 100:
		return enum.TierPremium

	default:
		return enum.InvalidTier
	}
}
