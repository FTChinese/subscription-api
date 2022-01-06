package subrepo

import (
	"github.com/FTChinese/subscription-api/pkg/ali"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/wechat"
	"github.com/FTChinese/subscription-api/test"
	"github.com/smartwalle/alipay"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestEnv_SaveWxPayload(t *testing.T) {
	env := New(db.MockMySQL(), zaptest.NewLogger(t))

	order := test.NewPersona().OrderBuilder().Build()

	type args struct {
		schema wechat.PayloadSchema
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Save webhook payload",
			args: args{
				schema: wechat.NewPayloadSchema(order.ID, test.NewWxWebhookPayload(order).ToMap()).WithKind(wechat.RowKindWebhook),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := env.SaveWxPayload(tt.args.schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("SaveWxPayload() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("Tests run")
		})
	}
}

func TestEnv_SaveAliWebhookPayload(t *testing.T) {

	env := New(db.MockMySQL(), zaptest.NewLogger(t))

	type args struct {
		p ali.WebhookPayload
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Save webhook payload",
			args: args{
				p: ali.NewWebhookPayload(
					&alipay.TradeNotification{}),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := env.SaveAliWebhookPayload(tt.args.p); (err != nil) != tt.wantErr {
				t.Errorf("SaveAliWebhookPayload() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
