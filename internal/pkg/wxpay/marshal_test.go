package wxpay

import (
	"fmt"
	"github.com/go-pay/gopay"
	"reflect"
	"testing"
)

func TestOrderParams(t *testing.T) {
	o := OrderParams{
		OrderParamsCommon: OrderParamsCommon{
			Body:        "buy standard",
			OutTradeNo:  "1234567890",
			TotalAmount: 29800,
			IP:          "",
			CallbackURL: "",
			TradeType:   "",
		},
		AppID:       "appid",
		MchID:       "mchid",
		DeviceInfo:  "",
		Nonce:       "nonce",
		Sign:        "sign",
		SignType:    "md5",
		Attach:      "",
		Currency:    "",
		StartTime:   "",
		EndTime:     "",
		OfferTag:    "",
		LimitPay:    "",
		ShowReceipt: false,
		ShareProfit: false,
	}

	values := reflect.ValueOf(o)
	typeOfV := values.Type()

	t.Logf("Type NumField %d", typeOfV.NumField())
	t.Logf("Value NumField %d", values.NumField())
	t.Logf("VisibleFields %d", len(reflect.VisibleFields(typeOfV)))

	for i, sf := range reflect.VisibleFields(reflect.TypeOf(o)) {
		fmt.Printf("Index %d, Name: %s, Type: %s, Kind: %s, Anonymous: %t\n", i, sf.Name, sf.Type, sf.Type.Kind(), sf.Anonymous)
	}

	for i := 0; i < typeOfV.NumField(); i++ {
		field := typeOfV.Field(i)
		fmt.Printf("Index %d, Name: %s, Type: %s, Anonymous %t\n", i, field.Name, field.Type, field.Anonymous)
	}
}

func Test_getTag(t *testing.T) {
	type args struct {
		tag string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 bool
	}{
		{
			name: "Tag not omit empty",
			args: args{
				tag: "id",
			},
			want:  "id",
			want1: false,
		},
		{
			name: "Tag omit empty",
			args: args{
				tag: "id,omitempty",
			},
			want:  "id",
			want1: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := getTag(tt.args.tag)
			if got != tt.want {
				t.Errorf("getTag() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("getTag() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestMarshal(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name string
		args args
		want gopay.BodyMap
	}{
		{
			name: "Marshal struct to map",
			args: args{
				v: OrderParams{
					OrderParamsCommon: OrderParamsCommon{
						Body:        "Standard/Year",
						OutTradeNo:  "FT1234567890",
						TotalAmount: 29800,
						IP:          "127.0.0.1",
						CallbackURL: "https://www.ftchinse.com/webhook/wxpay",
						TradeType:   "APP",
					},
					AppID:       "appid",
					MchID:       "merchant_id",
					DeviceInfo:  "",
					Nonce:       "random_string",
					Sign:        "",
					SignType:    "",
					Attach:      "",
					Currency:    "",
					StartTime:   "",
					EndTime:     "",
					OfferTag:    "",
					LimitPay:    "",
					ShowReceipt: false,
					ShareProfit: false,
				},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Marshal(tt.args.v)

			t.Logf("%+v", got)
		})
	}
}
