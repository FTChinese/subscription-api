package accounts

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/test"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestEnv_DeleteAccount(t *testing.T) {
	repo := test.NewRepo()

	a1 := test.NewPersona().EmailOnlyAccount()
	repo.MustCreateFtcAccount(a1)

	env := New(db.MockMySQL(), zaptest.NewLogger(t))

	type args struct {
		d account.DeletedUser
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Delete account",
			args: args{
				d: account.DeletedUser{
					ID:         a1.FtcID,
					Email:      a1.Email,
					CreatedUTC: chrono.TimeNow(),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := env.DeleteAccount(tt.args.d); (err != nil) != tt.wantErr {
				t.Errorf("DeleteAccount() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
