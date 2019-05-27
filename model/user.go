package model

import (
	"gitlab.com/ftchinese/subscription-api/paywall"
)

// FindUser retrieves user's name and email to be used in an email.
func (env Env) FindUser(id string) (paywall.FtcUser, error) {
	query := `
	SELECT user_id AS userId,
		wx_union_id AS unionId,
		user_name AS userName,
		email
	FROM cmstmp01.userinfo
	WHERE user_id = ?`

	var u paywall.FtcUser
	err := env.db.QueryRow(query, id).Scan(
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
