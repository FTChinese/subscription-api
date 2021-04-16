package accounts

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/internal/repository/readers"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/test"
	"github.com/brianvoe/gofakeit/v5"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestEnv_UpdateEmail(t *testing.T) {

	a := account.NewMockFtcAccountBuilder(enum.AccountKindFtc).Build()

	t.Logf("%s : %s", a.FtcID, a.Email)

	test.NewRepo().MustCreateFtcAccount(a)

	faker.SeedGoFake()

	a.Email = gofakeit.Email()

	type fields struct {
		Env readers.Env
	}
	type args struct {
		a account.BaseAccount
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Update email",
			fields: fields{
				Env: readers.New(test.SplitDB, zaptest.NewLogger(t)),
			},
			args: args{
				a: a,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				Env: tt.fields.Env,
			}
			if err := env.UpdateEmail(tt.args.a); (err != nil) != tt.wantErr {
				t.Errorf("UpdateEmail() error = %v, wantErr %v", err, tt.wantErr)
			}

			t.Logf("Changed to %s", tt.args.a.Email)
		})
	}
}

func TestEnv_SaveEmailHistory(t *testing.T) {
	a := account.NewMockFtcAccountBuilder(enum.AccountKindFtc).Build()

	t.Logf("%s : %s", a.FtcID, a.Email)

	test.NewRepo().MustCreateFtcAccount(a)

	type fields struct {
		Env readers.Env
	}
	type args struct {
		a account.BaseAccount
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Save email change history",
			fields: fields{
				Env: readers.New(test.SplitDB, zaptest.NewLogger(t)),
			},
			args: args{
				a: a,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				Env: tt.fields.Env,
			}
			if err := env.SaveEmailHistory(tt.args.a); (err != nil) != tt.wantErr {
				t.Errorf("SaveEmailHistory() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
