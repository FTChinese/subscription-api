package reader

import (
	"github.com/guregu/null"
	"strings"
)

// Account contains the minimal data to identify a user.
type Account struct {
	FtcID    string      `json:"ftcId" db:"ftc_id"`
	UnionID  null.String `json:"unionId" db:"union_id"`
	StripeID null.String `json:"stripeId" db:"stripe_id"`
	UserName null.String `json:"userName" db:"user_name"`
	Email    string      `json:"email" db:"email"`
}

func (a Account) MemberID() MemberID {
	return MemberID{
		CompoundID: "",
		FtcID:      null.NewString(a.FtcID, a.FtcID != ""),
		UnionID:    a.UnionID,
	}.MustNormalize()
}

func (a Account) IsSandbox() bool {
	return strings.HasSuffix(a.Email, ".sandbox@ftchinese.com")
}

// NormalizeName returns user name, or the name part of email if name does not exist.
func (a Account) NormalizeName() string {
	if a.UserName.Valid {
		return strings.Split(a.UserName.String, "@")[0]
	}

	return strings.Split(a.Email, "@")[0]
}
