package accounts

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/test"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestEnv_IDExists(t *testing.T) {

	env := New(test.SplitDB, zaptest.NewLogger(t))

	a := account.NewMockFtcAccountBuilder(enum.AccountKindFtc).Build()

	test.NewRepo().MustCreateFtcAccount(a)

	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "",
			args: args{
				id: a.FtcID,
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := env.IDExists(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("IDExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IDExists() got = %v, want %v", got, tt.want)
			}
		})
	}
}
