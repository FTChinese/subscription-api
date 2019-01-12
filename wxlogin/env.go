package wxlogin

import (
	"database/sql"
	"time"

	"gitlab.com/ftchinese/subscription-api/enum"
	"gitlab.com/ftchinese/subscription-api/postoffice"
)

// Env wraps database operations.
type Env struct {
	DB      *sql.DB
	PostMan postoffice.Postman
}

// Format Unix timestamp to a date string `2019-07-20`
func normalizeExpireDate(timestamp int64) time.Time {
	return time.Unix(timestamp, 0)
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
