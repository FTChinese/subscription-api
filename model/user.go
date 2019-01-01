package model

import (
	"html/template"
	"strings"

	"github.com/guregu/null"
	"gitlab.com/ftchinese/subscription-api/util"
)

// User contains the minimal information to identify a user.
type User struct {
	ID      string
	UnionID null.String
	Name    string
	Email   string
}

// NormalizeName returns user name, or the name part of email if name does not exist.
func (u User) NormalizeName() string {
	return strings.Split(u.Name, "@")[0]
}

// FindUser retrieves user's name and email to be used in an email.
func (env Env) FindUser(id string) (User, error) {
	query := `
	SELECT user_id AS id,
		user_name AS name,
		email
	FROM cmstmp01.userinfo
	WHERE user_id = ?`

	var u User
	err := env.DB.QueryRow(query, id).Scan(
		&u.ID,
		&u.Name,
		&u.Email,
	)

	if err != nil {
		logger.WithField("location", "FindUser").Error(err)

		return u, err
	}

	return u, nil
}

// ComposeEmail compiles letter templates against data.
func ComposeEmail(u User, s Subscription) (util.Parcel, error) {
	tmpl, err := template.New("order").Parse(letter)

	if err != nil {
		logger.WithField("location", "ComposeEmail").Error(err)
		return util.Parcel{}, err
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
		logger.WithField("location", "ComposeEmail").Error(err)
		return util.Parcel{}, err
	}

	return util.Parcel{
		FromAddress: "no-reply@ftchinese.com",
		FromName:    "FT中文网会员订阅",
		ToAddress:   u.Email,
		ToName:      u.NormalizeName(),
		Subject:     "会员订阅",
		Body:        body.String(),
	}, nil
}
