package accounts

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/test"
	"github.com/brianvoe/gofakeit/v5"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestEnv_SaveEmailVerifier(t *testing.T) {

	faker.SeedGoFake()

	env := New(db.MockMySQL(), zaptest.NewLogger(t))

	type args struct {
		v account.EmailVerifier
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Save email verification token",
			args: args{
				v: account.MustNewEmailVerifier(gofakeit.Email(), ""),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := env.SaveEmailVerifier(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("SaveEmailVerifier() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnv_RetrieveEmailVerifier(t *testing.T) {
	faker.SeedGoFake()

	v := account.MustNewEmailVerifier(gofakeit.Email(), "")

	test.NewRepo().MustSaveEmailVerifier(v)

	env := New(db.MockMySQL(), zaptest.NewLogger(t))

	type args struct {
		token string
	}
	tests := []struct {
		name    string
		args    args
		want    account.EmailVerifier
		wantErr bool
	}{
		{
			name: "Retrieve email verifier",
			args: args{
				token: v.Token,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := env.RetrieveEmailVerifier(tt.args.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("RetrieveEmailVerifier() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("RetrieveEmailVerifier() got = %v, want %v", got, tt.want)
			//}

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}

func TestEnv_EmailVerified(t *testing.T) {

	a := account.NewMockFtcAccountBuilder(enum.AccountKindFtc).Build()

	t.Logf("%s : %s", a.FtcID, a.Email)

	test.NewRepo().MustCreateFtcAccount(a)

	env := New(db.MockMySQL(), zaptest.NewLogger(t))

	type args struct {
		ftcID string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Flag email verified",
			args: args{
				ftcID: a.FtcID,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := env.EmailVerified(tt.args.ftcID); (err != nil) != tt.wantErr {
				t.Errorf("EmailVerified() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
