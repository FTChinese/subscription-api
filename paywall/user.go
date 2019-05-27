package paywall

import (
	"github.com/pkg/errors"
	"strings"
	"text/template"

	"github.com/FTChinese/go-rest/postoffice"
	"github.com/guregu/null"
)

// User is used to identify an FTC user.
// A user might have an ftc uuid, or a wechat union id,
// or both.
// This type structure is used to ensure unique constraint
// for SQL columns that cannot be both null since SQL do not
// have a mechanism to do UNIQUE INDEX on two columns while
// keeping either of them nullable.
// A user's compound id is taken from either ftc uuid or
// wechat id, with ftc id taking precedence.
type User struct {
	CompoundID string
	FtcID      null.String
	UnionID    null.String
}

// NewUser creates a new User instance and select the correct CompoundID
func NewUser(ftcID null.String, unionID null.String) (User, error) {
	u := User{
		FtcID:   ftcID,
		UnionID: unionID,
	}

	if ftcID.Valid {
		u.CompoundID = ftcID.String
	} else if unionID.Valid {
		u.CompoundID = unionID.String
	} else {
		return u, errors.New("ftcID and unionID should not both be null")
	}

	return u, nil
}

// FtcUser represents a row retrieve from userinfo table.
type FtcUser struct {
	UserID   string
	UnionID  null.String
	Email    string
	UserName null.String
}

// NormalizeName returns user name, or the name part of email if name does not exist.
func (u FtcUser) NormalizeName() string {
	if u.UserName.Valid {
		return strings.Split(u.UserName.String, "@")[0]
	}

	return strings.Split(u.Email, "@")[0]
}

// ConfirmationParcel create a parcel for email after subscription is confirmed.
func (u FtcUser) ConfirmationParcel(s Subscription) (postoffice.Parcel, error) {
	tmpl, err := template.New("order").Parse(confirmationLetter)

	if err != nil {
		return postoffice.Parcel{}, err
	}

	data := struct {
		User FtcUser
		Subs Subscription
	}{
		u,
		s,
	}

	var body strings.Builder
	err = tmpl.Execute(&body, data)

	if err != nil {
		return postoffice.Parcel{}, err
	}

	return postoffice.Parcel{
		FromAddress: "no-reply@ftchinese.com",
		FromName:    "FT中文网会员订阅",
		ToAddress:   u.Email,
		ToName:      u.NormalizeName(),
		Subject:     "会员订阅",
		Body:        body.String(),
	}, nil
}
