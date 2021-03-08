package subrepo

import (
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg"
	"github.com/FTChinese/subscription-api/pkg/ali"
	"github.com/FTChinese/subscription-api/test"
	"github.com/brianvoe/gofakeit/v5"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestAliPayClient_CreateOrder(t *testing.T) {

	faker.SeedGoFake()

	client := NewAliPayClient(test.AliApp, zaptest.NewLogger(t))

	type args struct {
		or ali.OrderReq
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Desktop order",
			args: args{
				or: ali.OrderReq{
					Title:       "Desktop order",
					FtcOrderID:  pkg.MustOrderID(),
					TotalAmount: "258.00",
					WebhookURL:  gofakeit.URL(),
					TxKind:      ali.EntryDesktopWeb,
					ReturnURL:   gofakeit.URL(),
				},
			},
			wantErr: false,
		},
		{
			name: "Mobile order",
			args: args{
				or: ali.OrderReq{
					Title:       "Mobile order",
					FtcOrderID:  pkg.MustOrderID(),
					TotalAmount: "258.00",
					WebhookURL:  gofakeit.URL(),
					TxKind:      ali.EntryMobileWeb,
					ReturnURL:   gofakeit.URL(),
				},
			},
			wantErr: false,
		},
		{
			name: "Mobile order",
			args: args{
				or: ali.OrderReq{
					Title:       "App order",
					FtcOrderID:  pkg.MustOrderID(),
					TotalAmount: "258.00",
					WebhookURL:  gofakeit.URL(),
					TxKind:      ali.EntryApp,
					ReturnURL:   gofakeit.URL(),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := client.CreateOrder(tt.args.or)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateOrder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("%s", got)
		})
	}
}

func TestAliPayClient_QueryOrder(t *testing.T) {
	c := NewAliPayClient(test.AliApp, zaptest.NewLogger(t))

	resp, err := c.QueryOrder("FT8F0438FFE67C7443")

	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("%+v", resp)
}
