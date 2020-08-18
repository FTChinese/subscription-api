package subrepo

import (
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/test"
	"testing"
)

func TestSubEnv_PreviewUpgrade(t *testing.T) {
	profile := test.NewProfile()

	store := test.NewSubStore(profile)
	orders := store.MustRenewN(3)

	// To have upgrading balance, a user must have an existing standard membership,
	// and some valid orders.
	repo := test.NewRepo()
	repo.MustSaveMembership(store.MustGetMembership())
	repo.MustSaveRenewalOrders(orders)

	env := SubEnv{db: test.DB}

	type args struct {
		userID reader.MemberID
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Preview upgrade wallet",
			args: args{
				userID: profile.AccountID(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := env.PreviewUpgrade(tt.args.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("PreviewUpgrade() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("Payment intent: %+v", got)
		})
	}
}
