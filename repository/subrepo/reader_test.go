package subrepo

import (
	"gitlab.com/ftchinese/subscription-api/test"
	"testing"
)

func TestEnv_FindFtcUser(t *testing.T) {

	profile := test.NewProfile()

	store := test.NewSubStore(profile)
	test.NewRepo(store).
		MustCreateAccount()

	env := SubEnv{
		db: test.DB,
	}

	type args struct {
		ftcId string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Find FTC Account",
			args: args{
				ftcId: profile.FtcID,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := env.FindFtcUser(tt.args.ftcId)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindFtcUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("FTC Account: %+v", got)
		})
	}
}

func TestEnv_FindStripeCustomer(t *testing.T) {

	profile := test.NewProfile()

	store := test.NewSubStore(profile)

	test.NewRepo(store).MustCreateAccount()

	env := SubEnv{
		db: test.DB,
	}

	type args struct {
		cusID string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Retrieve Account by Stripe ID",
			args: args{
				cusID: profile.StripeID,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := env.FindStripeCustomer(tt.args.cusID)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindStripeCustomer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("Stripe Account: %+v", got)
		})
	}
}
