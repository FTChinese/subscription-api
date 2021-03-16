package subrepo

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/FTChinese/subscription-api/pkg/wechat"
	"github.com/FTChinese/subscription-api/test"
	"go.uber.org/zap/zaptest"
	"testing"

	"github.com/smartwalle/alipay"
)

func TestEnv_SaveAliNotification(t *testing.T) {
	p := test.NewPersona()

	env := Env{
		dbs:    test.SplitDB,
		logger: zaptest.NewLogger(t),
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
			args:    args{n: subs.MockAliNoti(p.NewOrder(enum.OrderKindCreate))},
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

func TestEnv_SavePrepayResp(t *testing.T) {
	client := NewWxPayClient(test.WxPayApp, zaptest.NewLogger(t))

	p := test.NewPersona()
	or := test.NewWxOrderUnsigned()

	env := Env{
		dbs:    test.SplitDB,
		logger: zaptest.NewLogger(t),
	}

	type args struct {
		resp wechat.OrderResp
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Save Prepay Response",
			args: args{
				resp: wechat.NewOrderResp(p.NewOrder(enum.OrderKindCreate).ID, client.MockOrderPayload(or)),
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

	client := NewWxPayClient(test.WxPayApp, zaptest.NewLogger(t))

	p := test.NewPersona()
	noti := test.NewWxWHUnsigned(p.NewOrder(enum.OrderKindCreate))

	env := Env{
		dbs:    test.SplitDB,
		logger: zaptest.NewLogger(t),
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
				n: wechat.NewNotification(client.MockWebhookPayload(noti)),
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
