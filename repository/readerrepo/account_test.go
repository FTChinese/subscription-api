package readerrepo

import (
	"github.com/FTChinese/subscription-api/test"
	"testing"
)

func TestReaderEnv_FindAccountByFtcID(t *testing.T) {
	profile := test.NewPersona()
	test.NewRepo().MustSaveAccount(profile.Account())

	env := Env{db: test.DB}

	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "Find account ftc id",
			args:    args{id: profile.FtcID},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := env.AccountByFtcID(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindAccountByFtcID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("Account: %+v", got)
		})
	}
}

func TestReaderEnv_FindAccountByStripeID(t *testing.T) {
	profile := test.NewPersona()
	test.NewRepo().MustSaveAccount(profile.Account())

	env := Env{db: test.DB}

	type args struct {
		cusID string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "Account by stripe",
			args:    args{cusID: profile.StripeID},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := env.AccountByStripeID(tt.args.cusID)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindAccountByStripeID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("Acocunt: %+v", got)
		})
	}
}
