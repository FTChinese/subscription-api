package iaprepo

import (
	"github.com/FTChinese/go-rest/enum"
	"gitlab.com/ftchinese/subscription-api/models/apple"
	"gitlab.com/ftchinese/subscription-api/models/reader"
	"gitlab.com/ftchinese/subscription-api/test"
	"testing"
	"time"
)

// A linked IAP requesting link again.
func TestIAPEnv_Link_Update(t *testing.T) {
	profile := test.NewProfile().SetPayMethod(enum.PayMethodApple)

	test.NewRepo().MustSaveMembership(profile.Membership())

}

func TestIAPEnv_Link(t *testing.T) {
	profile := test.NewProfile()

	// Create an existing ftc member in db.
	existingFtc := test.NewProfile().Membership()
	test.NewRepo().MustSaveMembership(existingFtc)

	// Create an expired ftc member in db.
	expiredFtc := test.
		NewProfile().
		SetExpireDate(
			time.Now().AddDate(0, -1, 0),
		).Membership()
	test.NewRepo().MustSaveMembership(expiredFtc)

	// Create an IAP member in db.
	existingIAPProfile := test.NewProfile()
	existingIAP := existingIAPProfile.SetPayMethod(enum.PayMethodApple).Membership()
	test.NewRepo().MustSaveMembership(existingIAP)

	env := IAPEnv{
		db: test.DB,
	}

	type args struct {
		s  apple.Subscription
		id reader.MemberID
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "New IAP to existing ftc",
			args: args{
				s:  test.NewProfile().IAPSubs(),
				id: existingFtc.MemberID,
			},
			wantErr: true,
		},
		{
			name: "New IAP to expired FTC",
			args: args{
				s:  test.NewProfile().IAPSubs(),
				id: expiredFtc.MemberID,
			},
			wantErr: false,
		},
		{
			name: "New IAP to existing IAP",
			args: args{
				s:  test.NewProfile().IAPSubs(),
				id: existingIAP.MemberID,
			},
			wantErr: true,
		},
		{
			name: "New IAP to empty FTC",
			args: args{
				s:  profile.IAPSubs(),
				id: profile.AccountID(),
			},
			wantErr: false,
		},
		{
			name: "Update",
			args: args{
				s:  existingIAPProfile.IAPSubs(),
				id: existingIAP.MemberID,
			},
			wantErr: false,
		},
		{
			name: "Existing IAP to new ftc might be cheat",
			args: args{
				s:  existingIAPProfile.IAPSubs(),
				id: test.NewProfile().AccountID(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, got1, err := env.Link(tt.args.s, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Link() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("Linked membership: %+v", got)

			t.Logf("Is initial link: %t", got1)
		})
	}
}
