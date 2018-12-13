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

// EmailLogin is used for email + password login.
type EmailLogin struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// VerifyFTCAccount verifies if an email exists, and if password for this email is correct.
func (env Env) VerifyFTCAccount(l EmailLogin) {

}
