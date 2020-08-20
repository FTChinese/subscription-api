package iaprepo

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/apple"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/test"
	"testing"
)

// A linked IAP requesting link again.
func TestIAPEnv_Link_MockNewFTC(t *testing.T) {
	profile := test.NewPersona().SetPayMethod(enum.PayMethodApple)

	t.Logf("FTC id: %s", profile.FtcID)
}

func TestIAPEnv_Link(t *testing.T) {
	profile := test.NewPersona()

	// Create an existing ftc member in db.
	existingFtc := test.NewPersona().Membership()
	test.NewRepo().MustSaveMembership(existingFtc)

	// Create an expired ftc member in db.
	expiredFtc := test.
		NewPersona().
		SetExpired(true).
		Membership()
	test.NewRepo().MustSaveMembership(expiredFtc)

	// Create an IAP member in db.
	existingIAPProfile := test.NewPersona()
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
				s:  test.NewPersona().IAPSubs(),
				id: existingFtc.MemberID,
			},
			wantErr: true,
		},
		{
			name: "New IAP to expired FTC",
			args: args{
				s:  test.NewPersona().IAPSubs(),
				id: expiredFtc.MemberID,
			},
			wantErr: false,
		},
		{
			name: "New IAP to existing IAP",
			args: args{
				s:  test.NewPersona().IAPSubs(),
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
				id: test.NewPersona().AccountID(),
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
