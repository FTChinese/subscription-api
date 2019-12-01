package subrepo

import (
	"gitlab.com/ftchinese/subscription-api/models/reader"
	"gitlab.com/ftchinese/subscription-api/repository/query"
	"gitlab.com/ftchinese/subscription-api/test"
	"testing"
)

func TestEnv_FindFtcUser(t *testing.T) {

	profile := test.NewProfile()

	account := profile.Account(reader.AccountKindFtc)
	test.NewRepo().SaveAccount(account)

	env := SubEnv{
		db:    test.DB,
		query: query.NewBuilder(false),
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
				ftcId: account.FtcID,
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

	account := profile.Account(reader.AccountKindFtc)
	test.NewRepo().SaveAccount(account)

	env := SubEnv{
		db:    test.DB,
		query: query.NewBuilder(false),
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
