package model

import (
	"testing"

	"github.com/guregu/null"
	"github.com/icrowley/fake"
	"gitlab.com/ftchinese/subscription-api/util"
)

func TestConfirmationParcel(t *testing.T) {
	user := NewUser()
	subs := user.subs()

	p, err := user.ComfirmationParcel(subs)
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Parcel: %+v\n", p)
}

func TestSendEmail(t *testing.T) {
	user := User{
		UserID:   "2ebb0cdb-dfa2-4169-af2f-4def525c612e",
		UserName: null.StringFrom(fake.UserName()),
		Email:    "neefrankie@163.com",
	}

	err := user.createUser()
	if err != nil && !util.IsAlreadyExists(err) {
		t.Error(err)
		return
	}

	subs := user.subs()

	err = devEnv.SendConfirmationLetter(subs)
	if err != nil {
		t.Error(err)
	}
}
