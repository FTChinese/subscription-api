package subs

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"reflect"
	"testing"
	"time"
)

func TestConfirmationParams_confirmNewOrRenewalOrder(t *testing.T) {
	orderCreate := NewMockOrderBuilder("").Build()
	payResultForCreate := MockNewPaymentResult(orderCreate)

	orderRenewal := NewMockOrderBuilder("").WithKind(enum.OrderKindRenew).Build()
	payResultForRenewal := MockNewPaymentResult(orderRenewal)

	now := time.Now()

	type fields struct {
		Payment PaymentResult
		Order   Order
		Member  reader.Membership
	}
	tests := []struct {
		name   string
		fields fields
		// Only compare the start and end date.
		want dt.DateRange
	}{
		{
			name: "Order for new subscription",
			fields: fields{
				Payment: payResultForCreate,
				Order:   orderCreate,
				Member:  reader.Membership{},
			},
			want: dt.DateRange{
				StartDate: chrono.DateFrom(payResultForCreate.ConfirmedUTC.Time),
				EndDate:   chrono.DateFrom(payResultForCreate.ConfirmedUTC.AddDate(1, 0, 1)),
			},
		},
		{
			name: "Order for renewal",
			fields: fields{
				Payment: payResultForRenewal,
				Order:   orderRenewal,
				Member: reader.Membership{
					ExpireDate: chrono.DateFrom(now.AddDate(0, 3, 0)),
				},
			},
			want: dt.DateRange{
				StartDate: chrono.DateFrom(now.AddDate(0, 3, 0)),
				EndDate:   chrono.DateFrom(now.AddDate(1, 3, 1)),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := ConfirmationParams{
				Payment: tt.fields.Payment,
				Order:   tt.fields.Order,
				Member:  tt.fields.Member,
			}
			if got := params.confirmNewOrRenewalOrder(); !reflect.DeepEqual(got.DateRange, tt.want) {
				t.Errorf("confirmNewOrRenewalOrder() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfirmationParams_confirmUpgradeOrder(t *testing.T) {
	order := NewMockOrderBuilder("").WithKind(enum.OrderKindUpgrade).Build()
	payResult := MockNewPaymentResult(order)

	type fields struct {
		Payment PaymentResult
		Order   Order
		Member  reader.Membership
	}
	tests := []struct {
		name   string
		fields fields
		want   dt.DateRange
	}{
		{
			name: "Confirm upgrading order",
			fields: fields{
				Payment: payResult,
				Order:   order,
				Member: reader.Membership{
					ExpireDate: chrono.DateFrom(time.Now().AddDate(0, 3, 0)),
				},
			},
			want: dt.DateRange{
				StartDate: chrono.DateFrom(payResult.ConfirmedUTC.Time),
				EndDate:   chrono.DateFrom(payResult.ConfirmedUTC.AddDate(1, 0, 1)),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := ConfirmationParams{
				Payment: tt.fields.Payment,
				Order:   tt.fields.Order,
				Member:  tt.fields.Member,
			}
			if got := params.confirmUpgradeOrder(); !reflect.DeepEqual(got.DateRange, tt.want) {
				t.Errorf("confirmUpgradeOrder() = %v, want %v", got, tt.want)
			}
		})
	}
}
