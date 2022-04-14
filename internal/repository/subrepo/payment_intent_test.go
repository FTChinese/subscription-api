package subrepo

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/subscription-api/internal/pkg/ftcpay"
	"github.com/FTChinese/subscription-api/pkg/ali"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/wechat"
	"github.com/brianvoe/gofakeit/v5"
	"github.com/guregu/null"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestEnv_SavePaymentIntent(t *testing.T) {

	env := New(db.MockMySQL(), zaptest.NewLogger(t))

	type args struct {
		pi ftcpay.PaymentIntentSchema
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Save alipay intent",
			args: args{
				pi: ftcpay.PaymentIntentSchema{
					OrderID: ids.MustOrderID(),
					Price: price.FtcPriceJSON{
						FtcPrice: price.MockFtcStdYearPrice,
					},
					Offer:       price.DiscountColumn{},
					Membership:  reader.MembershipColumn{},
					WxPayParams: wechat.ColumnSDKParams{},
					AliPayParams: ali.ColumnSDKParams{
						SDKParams: ali.SDKParams{
							BrowserRedirect: null.StringFrom(gofakeit.URL()),
							AppSDK:          null.String{},
						},
					},
					CreatedUTC: chrono.TimeNow(),
				},
			},
		},
		{
			name: "Save wxpay intent",
			args: args{
				pi: ftcpay.PaymentIntentSchema{
					OrderID: ids.MustOrderID(),
					Price: price.FtcPriceJSON{
						FtcPrice: price.MockFtcStdYearPrice,
					},
					Offer:      price.DiscountColumn{},
					Membership: reader.MembershipColumn{},
					WxPayParams: wechat.ColumnSDKParams{
						SDKParams: wechat.SDKParams{
							DesktopQr:      null.String{},
							MobileRedirect: null.String{},
							JsApi:          wechat.JSApiParamsJSON{},
							AppSDK: wechat.NativeAppParamsJSON{
								NativeAppParams: wechat.NativeAppParams{
									AppID:     "test app id",
									PartnerID: "test partner id",
									PrepayID:  "tesst prepay id",
									Timestamp: "1234567890",
									Nonce:     "random string",
									Package:   "package",
									Signature: "signature",
								},
							},
						},
					},
					AliPayParams: ali.ColumnSDKParams{},
					CreatedUTC:   chrono.TimeNow(),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := env.SavePaymentIntent(tt.args.pi)

			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestEnv_RetrieveOrderPrice(t *testing.T) {
	env := New(db.MockMySQL(), zaptest.NewLogger(t))

	type args struct {
		orderID string
	}
	tests := []struct {
		name    string
		args    args
		want    price.FtcPriceJSON
		wantErr bool
	}{
		{
			name: "Retrieve price",
			args: args{
				orderID: "FT838E3D4DD90B04A6",
			},
			want:    price.FtcPriceJSON{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := env.RetrieveOrderPrice(tt.args.orderID)
			if (err != nil) != tt.wantErr {
				t.Errorf("RetrieveOrderPrice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("RetrieveOrderPrice() got = %v, want %v", got, tt.want)
			//}

			t.Logf("%v", got)
		})
	}
}
