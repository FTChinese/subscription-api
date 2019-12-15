package subrepo

import (
	"gitlab.com/ftchinese/subscription-api/models/subscription"
	"gitlab.com/ftchinese/subscription-api/test"
	"testing"
)

func TestSubEnv_FreeUpgrade(t *testing.T) {
	profile := test.NewProfile()

	store := test.NewSubStore(profile)

	// To have upgrading balance, a user must have an existing standard membership,
	// and some valid orders.
	repo := test.NewRepo()
	repo.MustSaveMembership(store.MustGetMembership())
	repo.MustSaveRenewalOrders(profile.StandardOrdersN(10))

	builder := subscription.NewOrderBuilder(profile.AccountID()).
		SetPlan(test.YearlyPremium).
		SetClient(test.RandomClientApp()).
		SetEnvironment(false)

	env := SubEnv{db: test.DB}

	type args struct {
		builder *subscription.OrderBuilder
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Free upgrade",
			args: args{
				builder: builder,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := env.FreeUpgrade(tt.args.builder)
			if (err != nil) != tt.wantErr {
				t.Errorf("FreeUpgrade() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("Confirmation result %+v", got)
		})
	}
}
