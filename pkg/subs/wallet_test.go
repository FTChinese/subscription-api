package subs

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/product"
	"github.com/guregu/null"
	"reflect"
	"testing"
	"time"
)

func TestNewWallet(t *testing.T) {
	type args struct {
		orders []BalanceSource
		asOf   time.Time
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Wallet",
			args: args{
				orders: []BalanceSource{
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

			for _, v := range got.Sources {
				t.Logf("%+v", v)
			}
		})
	}
}

func TestWallet_CheckOut(t *testing.T) {
	type fields struct {
		Balance   float64
		CreatedAt chrono.Time
		Sources   []ProratedOrder
	}
	type args struct {
		p product.ExpandedPlan
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   Checkout
	}{
		{
			name:   "Checkout without balance",
			fields: fields{},
			args: args{
				p: faker.PlanStdYear,
			},
			want: Checkout{
				Charge: product.Charge{
					Amount:     128,
					DiscountID: null.StringFrom("dsc_F7gEwjaF3OsR"),
					Currency:   "cny",
				},
				Duration: product.Duration{
					CycleCount: 1,
					ExtraDays:  1,
				},
			},
		},
		{
			name: "Checkout with balance",
			fields: fields{
				Balance: 100,
			},
			args: args{
				p: faker.PlanPrm,
			},
			want: Checkout{
				Charge: product.Charge{
					Amount:     1898,
					DiscountID: null.String{},
					Currency:   "cny",
				},
				Duration: product.Duration{
					CycleCount: 1,
					ExtraDays:  1,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := Wallet{
				Balance:   tt.fields.Balance,
				CreatedAt: tt.fields.CreatedAt,
				Sources:   tt.fields.Sources,
			}
			if got := w.CheckOut(tt.args.p); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CheckOut() = %v, want %v", got, tt.want)
			}
		})
	}
}
