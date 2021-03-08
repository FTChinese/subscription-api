package reader

import (
	"github.com/FTChinese/subscription-api/pkg"
	"github.com/guregu/null"
	"strings"
)

// FtcAccount contains the minimal data to identify a user.
type FtcAccount struct {
	FtcID    string      `json:"ftcId" db:"ftc_id"`
	UnionID  null.String `json:"unionId" db:"union_id"`
	StripeID null.String `json:"stripeId" db:"stripe_id"`
	Email    string      `json:"email" db:"email"`
	UserName null.String `json:"userName" db:"user_name"`
	VIP      bool        `json:"vip" db:"is_vip"`
}

func (a FtcAccount) MemberID() pkg.MemberID {
	return pkg.MemberID{
		CompoundID: "",
		FtcID:      null.NewString(a.FtcID, a.FtcID != ""),
		UnionID:    a.UnionID,
	}.MustNormalize()
}

func (a FtcAccount) IsTest() bool {
	return strings.HasSuffix(a.Email, ".test@ftchinese.com")
}

func (a FtcAccount) ValidateEnv(liveEnv bool) string {
	if liveEnv {
		// Live environment.
		if a.IsTest() {
			return "Please use sandbox environment for test account."
		}

		return ""
	} else {
		// Sandbox environment.
		if !a.IsTest() {
			return "Please use production environment for live account"
		}

		return ""
	}
}

func (a FtcAccount) IsFtc() bool {
	return a.FtcID != ""
}

func (a FtcAccount) IsWxOnly() bool {
	return a.FtcID == "" && a.UnionID.Valid
}

// NormalizeName returns user name, or the name part of email if name does not exist.
func (a FtcAccount) NormalizeName() string {
	if a.UserName.Valid {
		return strings.Split(a.UserName.String, "@")[0]
	}

	return strings.Split(a.Email, "@")[0]
}

type Account struct {
	FtcAccount
	Membership Membership `json:"membership"`
}
