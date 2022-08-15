package config

import (
	"github.com/FTChinese/go-rest/enum"
	"testing"
)

func TestAliWxWebhookURL(t *testing.T) {
	type args struct {
		isProd bool
		method enum.PayMethod
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Alipay production",
			args: args{
				isProd: true,
				method: enum.PayMethodAli,
			},
			want: "https://www.ftacademy.cn/api/v6/webhook/alipay",
		},
		{
			name: "Wechat production",
			args: args{
				isProd: true,
				method: enum.PayMethodWx,
			},
			want: "https://www.ftacademy.cn/api/v6/webhook/wxpay",
		},
		{
			name: "Alipay sandbox",
			args: args{
				isProd: false,
				method: enum.PayMethodAli,
			},
			want: "https://www.ftacademy.cn/api/sandbox/webhook/alipay",
		},
		{
			name: "Wechat sandbox",
			args: args{
				isProd: false,
				method: enum.PayMethodWx,
			},
			want: "https://www.ftacademy.cn/api/sandbox/webhook/wxpay",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AliWxWebhookURL(tt.args.isProd, tt.args.method); got != tt.want {
				t.Errorf("AliWxWebhookURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
