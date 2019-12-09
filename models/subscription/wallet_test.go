package subscription

import (
	"github.com/FTChinese/go-rest/chrono"
	"testing"
	"time"
)

func TestNewWallet(t *testing.T) {
	type args struct {
		orders []ProratedOrder
		asOf   time.Time
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Wallet",
			args: args{
				orders: []ProratedOrder{
					{
						OrderID:   "",
						Amount:    258,
						StartDate: chrono.DateFrom(time.Now().AddDate(0, -6, 0)),
						EndDate:   chrono.DateFrom(time.Now().AddDate(0, 6, 0)),
					},
					{
						OrderID:   "",
						Amount:    258,
						StartDate: chrono.DateFrom(time.Now().AddDate(0, -3, 0)),
						EndDate:   chrono.DateFrom(time.Now().AddDate(0, 9, 0)),
					},
					{
						OrderID:   "",
						Amount:    258,
						StartDate: chrono.DateFrom(time.Now()),
						EndDate:   chrono.DateFrom(time.Now().AddDate(1, 0, 0)),
					},
				},
				asOf: time.Now(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewWallet(tt.args.orders, tt.args.asOf)

			t.Logf("Balance: %f", got.Balance)

			for _, v := range got.Source {
				t.Logf("%+v", v)
			}
		})
	}
}
