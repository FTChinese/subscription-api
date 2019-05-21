package model

import (
	"gitlab.com/ftchinese/subscription-api/test"
	"testing"

	"gitlab.com/ftchinese/subscription-api/wechat"
)

func TestEnv_SavePrepayResp(t *testing.T) {
	env := Env{
		db: test.DB,
	}

	subs := test.MyProfile.SubsRandom()

	t.Logf("Subs: %+v", subs)

	type args struct {
		orderID string
		p       wechat.UnifiedOrderResp
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Save Prepay Response",
			args: args{
				orderID: subs.OrderID,
				p:       test.WxPrepay(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := env.SavePrepayResp(tt.args.orderID, tt.args.p); (err != nil) != tt.wantErr {
				t.Errorf("Env.SavePrepayResp() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnv_SaveWxNotification(t *testing.T) {
	env := Env{
		db: test.DB,
	}

	subs := test.MyProfile.SubsRandom()

	type args struct {
		n wechat.Notification
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Save Wxpay Notification",
			args: args{
				n: test.WxNotification(subs.OrderID),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := env.SaveWxNotification(tt.args.n); (err != nil) != tt.wantErr {
				t.Errorf("Env.SaveWxNotification() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
