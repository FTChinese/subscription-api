package model

import (
	"database/sql"

	"gitlab.com/ftchinese/subscription-api/member"

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

func normalizeMemberTier(vipType int64) member.Tier {
	switch vipType {

	case 10:
		return member.TierStandard

	case 100:
		return member.TierPremium

	default:
		return member.TierFree
	}
}

func normalizeExpireDate(timestamp int64) string {
	if timestamp == 0 {
		return ""
	}

	return util.ToSQLDateUTC.FromUnix(timestamp)
}
