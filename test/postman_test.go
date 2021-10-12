//go:build !production
// +build !production

package test

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/ztsms"
	"github.com/guregu/null"
	"testing"
)

func TestMobileDerivedEmailAccount(t *testing.T) {
	a := NewPersona().MobileOnlyAccount()

	repo := NewRepo()

	repo.MustCreateUserInfo(a)

	t.Logf("%s", faker.MustMarshalIndent(a))
}

// Generate sms verification parameters for a mobile-only user.
func TestVerifySMSForMobileEmail(t *testing.T) {
	p := NewPersona()

	repo := NewRepo()

	repo.MustCreateUserInfo(p.MobileOnlyAccount())

	v := ztsms.NewVerifier(p.Mobile, null.String{})
	repo.MustSaveMobileVerifier(v)

	t.Logf("%s", faker.MustMarshalIndent(ztsms.VerifierParams{
		Mobile:      v.Mobile,
		Code:        v.Code,
		DeviceToken: null.String{},
	}))
}

func TestLinkMobile(t *testing.T) {
	a := account.NewMockFtcAccountBuilder(enum.AccountKindFtc).WithMobile("").Build()
	v := ztsms.NewVerifier(faker.GenPhone(), null.String{})

	repo := NewRepo()
	repo.MustCreateFtcAccount(a)
	repo.MustSaveMobileVerifier(v)

	t.Logf("%s", faker.MustMarshalIndent(a))
	t.Logf("%s", faker.MustMarshalIndent(v))
}

func TestUnsetMobile(t *testing.T) {
	a := account.NewMockFtcAccountBuilder(enum.AccountKindFtc).Build()

	NewRepo().MustCreateFtcAccount(a)

	t.Logf("%s", faker.MustMarshalIndent(a))
}
