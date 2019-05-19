package model

import (
	"gitlab.com/ftchinese/subscription-api/test"
	"testing"

	"github.com/smartwalle/alipay"
)

func TestEnv_SaveAliNotification(t *testing.T) {
	env := Env{
		db: test.DB,
	}

	type args struct {
		n alipay.TradeNotification
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "Save Ali Notification",
			args:    args{n: test.AliNoti()},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := env.SaveAliNotification(tt.args.n); (err != nil) != tt.wantErr {
				t.Errorf("Env.SaveAliNotification() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
