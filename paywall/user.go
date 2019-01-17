package paywall

import (
	"strings"
	"text/template"

	"github.com/FTChinese/go-rest/postoffice"
	"github.com/guregu/null"
)

// User contains the minimal information to identify a user.
type User struct {
	UserID   string
	UnionID  null.String
	UserName null.String
	Email    string
}

// NormalizeName returns user name, or the name part of email if name does not exist.
func (u User) NormalizeName() string {
	if u.UserName.Valid {
		return strings.Split(u.UserName.String, "@")[0]
	}

	return strings.Split(u.Email, "@")[0]
}

// ComfirmationParcel create a parcel for email after subscription is confirmed.
func (u User) ComfirmationParcel(s Subscription) (postoffice.Parcel, error) {
	tmpl, err := template.New("order").Parse(confirmationLetter)

	if err != nil {
		return postoffice.Parcel{}, err
	}

	data := struct {
		User User
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
