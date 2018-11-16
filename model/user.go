package model

import (
	"strings"

	validate "github.com/asaskevich/govalidator"
	"gitlab.com/ftchinese/subscription-api/util"
)

// User contains the minimal information to identify a user.
type User struct {
	Name  string
	Email string
}

func (u User) NormalizeName() string {
	if validate.IsEmail(u.Name) {
		return strings.Split(u.Name, "@")[0]
	}

	return u.Name
}

// FindUser retrieves user's name and email to be used in an email.
func (env Env) FindUser(id string) (User, error) {
	query := `
	SELECT user_name,
		email
	FROM cmstmp01.userinfo
	WHERE user_id = ?`

	var u User
	err := env.DB.QueryRow(query, id).Scan(
		&u.Name,
		&u.Email,
	)

	if err != nil {
		logger.WithField("location", "FindUser").Error(err)

		return u, err
	}

	return u, nil
}

func ComposeEmail(s Subscription, u User) util.Parcel {
	return util.Parcel{
		FromAddress: "no-reply@ftchinese.com",
		FromName:    "FT中文网会员订阅",
		ToAddress:   u.Email,
		ToName:      u.Name,
		Subject:     "会员订阅",
		Body:        "",
	}
}
