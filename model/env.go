package model

import (
	"database/sql"

	log "github.com/sirupsen/logrus"
	"gitlab.com/ftchinese/subscription-api/util"
)

var logger = log.WithField("package", "subscription-api.model")

// Env wraps database connection
type Env struct {
	DB *sql.DB
}

func normalizeMemberTier(vipType int64) MemberTier {
	switch vipType {

	case 10:
		return TierStandard

	case 100:
		return TierPremium

	default:
		return TierInvalid
	}
}

func normalizeExpireDate(timestamp int64) string {
	if timestamp == 0 {
		return ""
	}

	return util.SQLDateUTC.FromUnix(timestamp)
}
