package subs

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/guregu/null"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestOrder_Confirm(t *testing.T) {

	pi := NewPayment(account, planStdYear).WithAlipay()

	order, err := pi.BuildOrder(pi.Checkout(nil, enum.OrderKindCreate))
	if err != nil {
		t.Error(err)
	}

	type args struct {
		pr PaymentResult
		m  reader.Membership
	}
	tests := []struct {
		name    string
		order   Order
		args    args
		want    ConfirmationResult
		wantErr bool
	}{
		{
			name:  "Confirm",
			order: order,
			args: args{
				pr: PaymentResult{
					Amount:        null.IntFrom(12800),
					TransactionID: "1234",
					OrderID:       order.ID,
					PaidAt:        chrono.TimeNow(),
				},
				m: reader.Membership{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := tt.order.Confirm(tt.args.pr, tt.args.m)
			assert.NoError(t, err)
			assert.NotZero(t, got.Membership)
			assert.Zero(t, got.Snapshot)
		})
	}
}

func TestOrder_pickStartDate(t *testing.T) {
	pi := NewPayment(account, planStdYear).WithAlipay()

	order, err := pi.BuildOrder(pi.Checkout(nil, enum.OrderKindCreate))
	if err != nil {
		t.Error(err)
	}

	order.ConfirmedAt = chrono.TimeNow()

	type args struct {
		expireDate chrono.Date
	}
	tests := []struct {
		name  string
		order Order
		args  args
		want  chrono.Date
	}{
		{
			name:  "Use order confirmation date",
			order: order,
			args:  args{expireDate: chrono.Date{}},
			want:  chrono.DateFrom(order.ConfirmedAt.Time),
		},
		{
			name:  "Use expiration date",
			order: order,
			args:  args{expireDate: chrono.DateFrom(time.Now().AddDate(0, 0, 1))},
			want:  chrono.DateFrom(time.Now().AddDate(0, 0, 1)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := tt.order
			got := o.pickStartDate(tt.args.expireDate)
			assert.Equal(t, got, tt.want)
		})
	}
}
