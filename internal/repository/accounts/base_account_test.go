package accounts

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/test"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestEnv_CreateAccount(t *testing.T) {

	env := New(db.MockMySQL(), zaptest.NewLogger(t))

	type args struct {
		a account.BaseAccount
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Create ftc-only account",
			args: args{
				a: test.NewPersona().EmailOnlyAccount(),
			},
			wantErr: false,
		},
		{
			name: "Create mobile account",
			args: args{
				a: test.NewPersona().MobileOnlyAccount(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := env.CreateAccount(tt.args.a); (err != nil) != tt.wantErr {
				t.Errorf("CreateAccount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("Created account %s", tt.args.a.FtcID)
		})
	}
}

func TestEnv_BaseAccountByEmail(t *testing.T) {
	a := account.NewMockFtcAccountBuilder(enum.AccountKindFtc).Build()

	test.NewRepo().MustCreateFtcAccount(a)

	env := New(db.MockMySQL(), zaptest.NewLogger(t))

	type args struct {
		email string
	}
	tests := []struct {
		name    string
		args    args
		want    account.BaseAccount
		wantErr bool
	}{
		{
			name: "Retrieve base account by email",
			args: args{
				email: a.Email,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := env.BaseAccountByEmail(tt.args.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("BaseAccountByEmail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("BaseAccountByEmail() got = %v, want %v", got, tt.want)
			//}

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}
