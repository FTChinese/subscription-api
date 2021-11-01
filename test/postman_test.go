//go:build !production
// +build !production

package test

import (
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/ztsms"
	"github.com/guregu/null"
	"testing"
)

// Create a row in userinfo for mobile-only email.
func TestEmailSignUp_WithMobile(t *testing.T) {
	a := NewPersona().EmailMobileAccount()

	t.Logf("%s", faker.MustMarshalIndent(a))
}

func TestVerifySMSCode_NewMobile(t *testing.T) {
	repo := NewRepo()

	v := ztsms.NewVerifier(faker.GenPhone(), null.String{})
	repo.MustSaveMobileVerifier(v)

	t.Logf("%s", faker.MustMarshalIndent(v))
}

func TestVerifySMSCode_ExistingMobile(t *testing.T) {
	a := NewPersona().EmailMobileAccount()
	repo := NewRepo()

	repo.MustCreateFtcAccount(a)

	v := ztsms.NewVerifier(a.Mobile.String, null.StringFrom(a.FtcID))
	repo.MustSaveMobileVerifier(v)

	t.Logf("%s", faker.MustMarshalIndent(v))
}

// Generate sms verification parameters for a mobile-only user.
// * Create a mobile-derived account in userinfo.
// * Create a verifier for this phone.
func TestVerifySMSCode_MobileOnlyAccount(t *testing.T) {
	repo := NewRepo()

	a := NewPersona().MobileOnlyAccount()
	repo.MustCreateUserInfo(a)

	v := ztsms.NewVerifier(a.Mobile.String, null.String{})
	repo.MustSaveMobileVerifier(v)

	t.Logf("%s", faker.MustMarshalIndent(a))
	t.Logf("%s", faker.MustMarshalIndent(v))
}

func TestMobileSignUp_RealPhone(t *testing.T) {
	v := ztsms.NewVerifier("15011481214", null.String{})

	NewRepo().MustSaveMobileVerifier(v)

	t.Logf("%s", faker.MustMarshalIndent(v))
}

// Mobile login case 1: mobile used for the 1st time,
// and user wants to create a new account with this mobile.
// Simply generate a new phone works.
func TestMobileSignUp_NewMobileAccount(t *testing.T) {
	v := ztsms.NewVerifier(faker.GenPhone(), null.String{})

	NewRepo().MustSaveMobileVerifier(v)

	t.Logf("%s", faker.MustMarshalIndent(v))
}

// Mobile login case 2: mobile used for the 1st time,
// and user wants to create a new email account with
// mobile set.
// Generates a new phone, and generates a new email account
// but does not save it.
func TestMobileSignUp_LinkNewEmail(t *testing.T) {
	a := NewPersona().EmailOnlyAccount()

	v := ztsms.NewVerifier(faker.GenPhone(), null.String{})

	NewRepo().MustSaveMobileVerifier(v)

	t.Logf("%s", faker.MustMarshalIndent(a))
	t.Logf("%s", faker.MustMarshalIndent(v))
}

// Mobile login case 3: mobile used for the 1st time,
// and user wants to link to an existing email account.
// Generate a new phone, and create an email-only account.
func TestMobileSignUp_LinkExistingEmail(t *testing.T) {
	a := NewPersona().EmailOnlyAccount()
	v := ztsms.NewVerifier(faker.GenPhone(), null.String{})

	repo := NewRepo()
	repo.MustCreateFtcAccount(a)
	repo.MustSaveMobileVerifier(v)

	t.Logf("%s", faker.MustMarshalIndent(a))
	t.Logf("%s", faker.MustMarshalIndent(v))
}

func TestEmailVerification_NewToken(t *testing.T) {
	p := NewPersona()

	vrf, err := account.NewEmailVerifier(p.Email, "")
	if err != nil {
		panic(err)
	}

	repo := NewRepo()
	repo.CreateFtcAccount(p.EmailOnlyAccount())
	repo.MustSaveEmailVerifier(vrf)

	t.Logf("%s", faker.MustMarshalIndent(vrf))
}

func TestLoadAccountByFtcID_SyncMobile(t *testing.T) {
	repo := NewRepo()

	a := NewPersona().MobileOnlyAccount()
	repo.CreateUserInfo(a)

	t.Logf("%s", faker.MustMarshalIndent(a))
}

func TestAccount_DeleteNoMembership(t *testing.T) {
	a := NewPersona().EmailOnlyAccount()

	repo := NewRepo()

	repo.CreateFtcAccount(a)

	t.Logf("%s", faker.MustMarshalIndent(a))
}

func TestAccount_DeleteWithValidMembership(t *testing.T) {
	p := NewPersona()
	a := p.EmailOnlyAccount()
	m := p.MemberBuilder().Build()

	repo := NewRepo()

	repo.MustCreateFtcAccount(a)
	repo.MustSaveMembership(m)

	t.Logf("%s", faker.MustMarshalIndent(a))
	t.Logf("%s", faker.MustMarshalIndent(m))
}

func TestAccount_ChangePassword(t *testing.T) {
	a := NewPersona().EmailOnlyAccount()

	repo := NewRepo()

	repo.CreateFtcAccount(a)

	t.Logf("%s", faker.MustMarshalIndent(a))
}
