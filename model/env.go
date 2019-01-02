package model

import (
	"database/sql"
	"time"

	"gitlab.com/ftchinese/subscription-api/enum"

	"github.com/patrickmn/go-cache"

	log "github.com/sirupsen/logrus"
	"gitlab.com/ftchinese/subscription-api/util"
)

var logger = log.WithField("package", "subscription-api.model")

const (
	keySchedule = "discountSchedule"
	keyPromo    = "promotionSchedule"
)

// Env wraps database connection
type Env struct {
	DB         *sql.DB
	Cache      *cache.Cache
	PostOffice util.PostOffice
}

func normalizeMemberTier(vipType int64) enum.Tier {
	switch vipType {

	case 10:
		return enum.TierStandard

	case 100:
		return enum.TierPremium

	default:
		return enum.TierFree
	}
}

func normalizeExpireDate(timestamp int64) time.Time {
	return time.Unix(timestamp, 0)
}
