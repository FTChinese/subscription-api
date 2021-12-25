package wechat

import (
	"github.com/guregu/null"
	"github.com/objcoding/wxpay"
	"testing"
)

func TestUnifiedOrderReq_Marshal(t *testing.T) {
	type fields struct {
		UnifiedOrderParams UnifiedOrderParams
		AppID              string
		MchID              string
		DeviceInfo         string
		Nonce              string
		Sign               string
		SignType           string
		Attach             string
		Currency           string
		StartTime          string
		EndTime            string
		OfferTag           string
		LimitPay           string
		ShowReceipt        bool
		ShareProfit        bool
	}
	tests := []struct {
		name   string
		fields UnifiedOrderReq
		want   wxpay.Params
	}{
		{
			name: "Marshal to map",
			fields: UnifiedOrderReq{
				UnifiedOrderParams: UnifiedOrderParams{
					Body:        "Standard/Year",
					OutTradeNo:  "FT12345678",
					TotalAmount: 29800,
					UserIP:      "127.0.0.1",
					WebhookURL:  "https://www.example.org",
					TradeType:   "APP",
					OpenID:      null.String{},
				},
				AppID:       "",
				MchID:       "",
				DeviceInfo:  "",
				Nonce:       GenerateNonce(),
				Sign:        "",
				SignType:    "",
				Attach:      "",
				Currency:    "",
				StartTime:   "",
				EndTime:     "",
				OfferTag:    "",
				LimitPay:    "",
				ShowReceipt: true,
				ShareProfit: false,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := tt.fields
			got := o.Marshal()

			t.Logf("%+v", got)
		})
	}
}
