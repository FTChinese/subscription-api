package model

import (
	"github.com/guregu/null"
	"gitlab.com/ftchinese/subscription-api/paywall"
	"gitlab.com/ftchinese/subscription-api/test"
	"testing"
)

func TestConfirmationParcel(t *testing.T) {
	subs := test.MyProfile.ConfirmedSubs()

	ftcUser := paywall.FtcUser{
		UserID:   test.MyProfile.FtcID,
		Email:    test.MyProfile.Email,
		UserName: null.StringFrom(test.MyProfile.UserName),
	}

	p, err := ftcUser.ConfirmationParcel(subs)
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Parcel: %+v\n", p)
}

func TestSendEmail(t *testing.T) {

	ftcUser := test.MyProfile.FtcUser()

	subs := test.MyProfile.ConfirmedSubs()

	parcel, err := ftcUser.ConfirmationParcel(subs)

	if err != nil {
		t.Error(err)
		return
	}

	err = test.Postman.Deliver(parcel)
	if err != nil {
		t.Error(err)
	}
}
