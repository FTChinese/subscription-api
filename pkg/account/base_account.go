package account

import (
	"github.com/FTChinese/subscription-api/pkg"
	"github.com/guregu/null"
	"strings"
)

// Ftc contains the minimal information to identify a user.
type BaseAccount struct {
	FtcID      string      `json:"id" db:"ftc_id"`           // FTC's uuid
	UnionID    null.String `json:"unionId" db:"wx_union_id"` // Wechat's union id
	StripeID   null.String `json:"stripeId" db:"stripe_id"`  // Stripe's id
	Email      string      `json:"email" db:"email"`         // Required, unique. Max 64.
	Password   string      `json:"-" db:"password"`
	Mobile     null.String `json:"mobile" db:"mobile_phone"`
	UserName   null.String `json:"userName" db:"user_name"` // Required, unique. Max 64.
	AvatarURL  null.String `json:"avatarUrl" db:"ftc_avatar_url"`
	IsVerified bool        `json:"isVerified" db:"is_verified"`
}

func (a BaseAccount) WithMobile(m string) BaseAccount {
	a.Mobile = null.StringFrom(m)
	return a
}

func (a BaseAccount) CompoundIDs() pkg.UserIDs {
	return pkg.UserIDs{
		CompoundID: "",
		FtcID:      null.NewString(a.FtcID, a.FtcID != ""),
		UnionID:    a.UnionID,
	}.MustNormalize()
}

func (a BaseAccount) IsTest() bool {
	return strings.HasSuffix(a.Email, ".test@ftchinese.com")
}

func (a BaseAccount) ValidateEnv(liveEnv bool) string {
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

func (a BaseAccount) IsFtc() bool {
	return a.FtcID != ""
}

func (a BaseAccount) IsWxOnly() bool {
	return a.FtcID == "" && a.UnionID.Valid
}

// NormalizeName returns user name, or the name part of email if name does not exist.
func (a BaseAccount) NormalizeName() string {
	if a.UserName.Valid {
		return strings.Split(a.UserName.String, "@")[0]
	}

	return strings.Split(a.Email, "@")[0]
}
