package accounts

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/test"
	"github.com/brianvoe/gofakeit/v5"
	"github.com/guregu/null"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestEnv_UpdateUserName(t *testing.T) {
	faker.SeedGoFake()

	a := account.NewMockFtcAccountBuilder(enum.AccountKindFtc).Build()

	test.NewRepo().MustCreateFtcAccount(a)

	t.Logf("%s", a.FtcID)

	a.UserName = null.StringFrom(gofakeit.Username())

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
			name: "Update user name",
			args: args{
				a: a,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := env.UpdateUserName(tt.args.a); (err != nil) != tt.wantErr {
				t.Errorf("UpdateUserName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
