package ali

import (
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/test"
	"github.com/brianvoe/gofakeit/v5"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestAliPayClient_CreateOrder(t *testing.T) {

	faker.SeedGoFake()

	client := NewPayClient(test.AliApp, zaptest.NewLogger(t))

	type args struct {
		or OrderReq
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Desktop order",
			args: args{
				or: OrderReq{
					Title:       "Desktop order",
					FtcOrderID:  ids.MustOrderID(),
					TotalAmount: "258.00",
					WebhookURL:  gofakeit.URL(),
					TxKind:      EntryDesktopWeb,
					ReturnURL:   gofakeit.URL(),
				},
			},
			wantErr: false,
		},
		{
			name: "Mobile order",
			args: args{
				or: OrderReq{
					Title:       "Mobile order",
					FtcOrderID:  ids.MustOrderID(),
					TotalAmount: "258.00",
					WebhookURL:  gofakeit.URL(),
					TxKind:      EntryMobileWeb,
					ReturnURL:   gofakeit.URL(),
				},
			},
			wantErr: false,
		},
		{
			name: "Mobile order",
			args: args{
				or: OrderReq{
					Title:       "App order",
					FtcOrderID:  ids.MustOrderID(),
					TotalAmount: "258.00",
					WebhookURL:  gofakeit.URL(),
					TxKind:      EntryApp,
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
	c := NewPayClient(test.AliApp, zaptest.NewLogger(t))

	resp, err := c.QueryOrder("FT8F0438FFE67C7443")

	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("%+v", resp)
}
