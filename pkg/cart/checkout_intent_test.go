package cart

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"reflect"
	"testing"
)

func Test_formatMethods(t *testing.T) {
	type args struct {
		methods []enum.PayMethod
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Stringify an array of one payment methods",
			args: args{
				methods: []enum.PayMethod{
					enum.PayMethodAli,
				},
			},
			want: "alipay",
		},
		{
			name: "Stringify an array of payment methods",
			args: args{
				methods: []enum.PayMethod{
					enum.PayMethodAli,
					enum.PayMethodWx,
					enum.PayMethodStripe,
				},
			},
			want: "alipay, wechat or stripe",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatMethods(tt.args.methods); got != tt.want {
				t.Errorf("formatMethods() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckoutIntents_Get(t *testing.T) {
	type fields struct {
		intents []CheckoutIntent
		err     error
	}
	type args struct {
		m enum.PayMethod
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    CheckoutIntent
		wantErr bool
	}{
		{
			name: "Find intent by payment method",
			fields: fields{
				intents: []CheckoutIntent{
					NewOneTimeIntent(enum.OrderKindCreate),
					NewSubsIntent(SubsKindNew),
				},
				err: nil,
			},
			args: args{
				m: enum.PayMethodStripe,
			},
			want: CheckoutIntent{
				OneTimeKind: enum.OrderKindNull,
				SubsKind:    SubsKindNew,
				PayMethods: []enum.PayMethod{
					enum.PayMethodStripe,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coi := reader.CheckoutIntents{
				intents: tt.fields.intents,
				err:     tt.fields.err,
			}
			got, err := coi.Get(tt.args.m)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get() got = %v, want %v", got, tt.want)
			}
		})
	}
}
