package model

import (
	"github.com/guregu/null"
	"gitlab.com/ftchinese/subscription-api/paywall"
	"gitlab.com/ftchinese/subscription-api/test"
	"testing"
)

func TestConfirmationParcel(t *testing.T) {
	p := test.MyProfile
	u := p.RandomUser()
	subs := test.SubsConfirmed(u)

	ftcUser := paywall.FtcUser{
		UserID:   test.MyProfile.FtcID,
		Email:    test.MyProfile.Email,
		UserName: null.StringFrom(test.MyProfile.UserName),
	}

	parcel, err := ftcUser.ConfirmationParcel(subs)
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Parcel: %+v\n", parcel)
}

func TestSendEmail(t *testing.T) {

	ftcUser := test.MyProfile.FtcUser()
	u := test.MyProfile.User(test.IDFtc)

	subs := test.SubsConfirmed(u)

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
