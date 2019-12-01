package subrepo

import (
	"gitlab.com/ftchinese/subscription-api/models/query"
	"gitlab.com/ftchinese/subscription-api/models/wechat"
	"gitlab.com/ftchinese/subscription-api/test"
	"testing"

	"github.com/smartwalle/alipay"
)

func TestEnv_SaveAliNotification(t *testing.T) {
	env := SubEnv{
		db:    test.DB,
		query: query.NewBuilder(false),
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
				t.Errorf("SubEnv.SaveAliNotification() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnv_SavePrepayResp(t *testing.T) {
	env := SubEnv{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	type args struct {
		resp wechat.UnifiedOrderResp
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Save Prepay Response",
			args: args{
				resp: test.WxPrepay(test.MustGenOrderID()),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := env.SavePrepayResp(tt.args.resp); (err != nil) != tt.wantErr {
				t.Errorf("SavePrepayResp() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnv_SaveWxNotification(t *testing.T) {
	env := SubEnv{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	type args struct {
		n wechat.Notification
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Save Wx Notification",
			args: args{
				n: test.WxNotification(test.MustGenOrderID()),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := env.SaveWxNotification(tt.args.n); (err != nil) != tt.wantErr {
				t.Errorf("SaveWxNotification() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
