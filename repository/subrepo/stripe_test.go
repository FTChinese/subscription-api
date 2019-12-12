package subrepo

import (
	"gitlab.com/ftchinese/subscription-api/test"
	"testing"
)

func TestSubEnv_CreateStripeCustomer(t *testing.T) {
	profile := test.NewProfile()

	test.NewRepo().MustCreateAccount(profile.Account())

	env := SubEnv{db: test.DB}

	account, err := env.CreateStripeCustomer(profile.FtcID)

	if err != nil {
		t.Error(err)
	}

	t.Logf("Account: %+v", account)
}
