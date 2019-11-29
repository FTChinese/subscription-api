package reader

import (
	"github.com/guregu/null"
	"strings"
)

// Account contains the minimal data to identify a user.
type Account struct {
	FtcID    string      `db:"ftc_id"`
	UnionID  null.String `db:"union_id"`
	StripeID null.String `db:"stripe_id"`
	Email    string      `db:"email"`
	UserName null.String `db:"user_name"`
}

func (a Account) ID() MemberID {
	id, _ := NewMemberID(a.FtcID, a.UnionID.String)
	return id
}

// NormalizeName returns user name, or the name part of email if name does not exist.
func (a Account) NormalizeName() string {
	if a.UserName.Valid {
		return strings.Split(a.UserName.String, "@")[0]
	}

	return strings.Split(a.Email, "@")[0]
}
