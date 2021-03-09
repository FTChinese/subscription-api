package reader

import (
	"github.com/FTChinese/go-rest/enum"
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
