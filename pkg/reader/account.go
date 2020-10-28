package reader

import (
	"github.com/guregu/null"
	"strings"
)

// Account contains the minimal data to identify a user.
type FtcAccount struct {
	FtcID    string      `json:"ftcId" db:"ftc_id"`
	UnionID  null.String `json:"unionId" db:"union_id"`
	StripeID null.String `json:"stripeId" db:"stripe_id"`
	Email    string      `json:"email" db:"email"`
	UserName null.String `json:"userName" db:"user_name"`
	VIP      bool        `json:"vip" db:"is_vip"`
}

func (a FtcAccount) MemberID() MemberID {
	return MemberID{
		CompoundID: "",
		FtcID:      null.NewString(a.FtcID, a.FtcID != ""),
		UnionID:    a.UnionID,
	}.MustNormalize()
}

func (a FtcAccount) IsTest() bool {
	return strings.HasSuffix(a.Email, ".test@ftchinese.com")
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
