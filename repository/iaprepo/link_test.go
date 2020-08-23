package iaprepo

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/apple"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/test"
	"testing"
)

func TestIAPEnv_Link(t *testing.T) {
	ftcP := test.NewPersona()

	repo := test.NewRepo()
	// Create an existing ftc member in db.
	validFtc := test.
		NewPersona().
		Membership()

	// Create an expired ftc member in db.
	expiredFtc := test.
		NewPersona().
		SetExpired(true).
		Membership()

	repo.MustSaveMembership(validFtc)
	repo.MustSaveMembership(expiredFtc)

	// Create an IAP member in db.
	iapP := test.NewPersona()
	iapMember := iapP.
		SetPayMethod(enum.PayMethodApple).
		Membership()
	repo.MustSaveMembership(iapMember)

	env := Env{
		cfg: test.CFG,
		db:  test.DB,
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
			name: "New IAP cannot link to existing valid ftc",
			args: args{
				s:  test.NewPersona().IAPSubs(),
				id: validFtc.MemberID,
			},
			wantErr: true,
		},
		{
			name: "New IAP link to expired FTC",
			args: args{
				s:  test.NewPersona().IAPSubs(),
				id: expiredFtc.MemberID,
			},
			wantErr: false,
		},
		{
			name: "New IAP cannot link to another existing IAP",
			args: args{
				s:  test.NewPersona().IAPSubs(),
				id: iapMember.MemberID,
			},
			wantErr: true,
		},
		{
			name: "New IAP can link to an empty FTC",
			args: args{
				s:  ftcP.IAPSubs(),
				id: ftcP.AccountID(),
			},
			wantErr: false,
		},
		{
			name: "Update",
			args: args{
				s:  iapP.IAPSubs(),
				id: iapMember.MemberID,
			},
			wantErr: false,
		},
		{
			name: "Existing IAP to new ftc might be cheat",
			args: args{
				s:  iapP.IAPSubs(),
				id: test.NewPersona().AccountID(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := env.Link(tt.args.s, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Link() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("Linked reader id: %s", got.Linked.CompoundID)
		})
	}
}
