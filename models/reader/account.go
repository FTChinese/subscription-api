package reader

import (
	"github.com/guregu/null"
	"strings"
)

// Account contains the minimal data to identify a user.
type Account struct {
	FtcID    string
	UnionID  null.String
	StripeID null.String
	Email    string
	UserName null.String
}

func (a Account) ID() AccountID {
	id, _ := NewID(a.FtcID, a.UnionID.String)
	return id
}

// NormalizeName returns user name, or the name part of email if name does not exist.
func (a Account) NormalizeName() string {
	if a.UserName.Valid {
		return strings.Split(a.UserName.String, "@")[0]
	}

	return strings.Split(a.Email, "@")[0]
}
