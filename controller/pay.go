package controller

import (
	"github.com/FTChinese/go-rest/postoffice"
	"gitlab.com/ftchinese/subscription-api/model"
	"gitlab.com/ftchinese/subscription-api/paywall"
)

// PayRouter is the base type used to handle shared payment operations.
type PayRouter struct {
	model   model.Env
	postman postoffice.Postman
}

// SendConfirmationLetter sends a confirmation email if user logged in with FTC account.
func (router PayRouter) sendConfirmationEmail(subs paywall.Subscription) error {
	if subs.IsWxLogin() {
		return nil
	}
	// 1. Find this user's personal data
	user, err := router.model.FindUser(subs.UserID)

	if err != nil {
		return err
	}

	parcel, err := user.ComfirmationParcel(subs)
	if err != nil {
		return err
	}

	logger.WithField("trace", "SendConirmationLetter").Info("Send subscription confirmation letter")

	err = router.postman.Deliver(parcel)
	if err != nil {
		logger.WithField("trace", "SendConfirmationLetter").Error(err)
		return err
	}
	return nil
}
