package model

import (
	"database/sql"
	"time"

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

func normalizeExpireTime(timestamp int64) string {
	if timestamp == 0 {
		return ""
	}

	return util.ISO8601UTC.FromUnix(timestamp)
}

func normalizeStartTime(timestamp int64) string {
	if timestamp == 0 {
		return ""
	}

	t := time.Unix(timestamp, 0).UTC().AddDate(-1, 0, 0)

	return util.ISO8601UTC.FromTime(t)
}
