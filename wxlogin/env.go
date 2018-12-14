package wxlogin

import (
	"database/sql"

	"gitlab.com/ftchinese/subscription-api/enum"
	"gitlab.com/ftchinese/subscription-api/util"
)

// Env wraps database operations.
type Env struct {
	DB *sql.DB
}

// Format Unix timestamp to a date string `2019-07-20`
func normalizeExpireDate(timestamp int64) string {
	if timestamp == 0 {
		return ""
	}

	return util.ToSQLDateUTC.FromUnix(timestamp)
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
