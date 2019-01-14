package model

import (
	"gitlab.com/ftchinese/subscription-api/paywall"
)

// findUser retrieves user's name and email to be used in an email.
func (env Env) findUser(id string) (paywall.User, error) {
	query := `
	SELECT user_id AS userId,
		wx_union_id AS unionId,
		user_name AS userName,
		email
	FROM cmstmp01.userinfo
	WHERE user_id = ?`

	var u paywall.User
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
func (env Env) SendConfirmationLetter(subs paywall.Subscription) error {
	if subs.IsWxLogin() {
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

	logger.WithField("trace", "SendConirmationLetter").Info("Send subscription confirmation letter")

	err = env.Postman.Deliver(parcel)
	if err != nil {
		logger.WithField("trace", "SendConfirmationLetter").Error(err)
		return err
	}
	return nil
}
