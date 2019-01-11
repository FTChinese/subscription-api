package model

import (
	"html/template"
	"strings"

	"github.com/guregu/null"
	"gitlab.com/ftchinese/subscription-api/postoffice"
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
	tmpl, err := template.New("order").Parse(postoffice.ConfirmationLetter)

	if err != nil {
		logger.WithField("trace", "ConfirmationParcel").Error(err)
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
		logger.WithField("trace", "ComposeParcel").Error(err)
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

// findUser retrieves user's name and email to be used in an email.
func (env Env) findUser(id string) (User, error) {
	query := `
	SELECT user_id AS userId,
		wx_union_id AS unionId,
		user_name AS userName,
		email
	FROM cmstmp01.userinfo
	WHERE user_id = ?`

	var u User
	err := env.DB.QueryRow(query, id).Scan(
		&u.UserID,
		&u.UnionID,
		&u.UserName,
		&u.Email,
	)

	if err != nil {
		logger.WithField("trace", "FindUser").Error(err)

		return u, err
	}

	return u, nil
}

// SendConfirmationLetter sends a confirmation email if user logged in with FTC account.
func (env Env) SendConfirmationLetter(subs Subscription) error {
	if subs.isWxLogin() {
		return nil
	}
	// 1. Find this user's personal data
	user, err := env.findUser(subs.UserID)

	if err != nil {
		return err
	}

	parcel, err := user.ComfirmationParcel(subs)
	if err != nil {
		return err
	}

	err = env.PostMan.Deliver(parcel)

	return err
}
