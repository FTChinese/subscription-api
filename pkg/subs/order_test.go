package subs

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/google/uuid"
	"testing"
	"time"
)

func TestOrder_IsSynced(t *testing.T) {

	now := time.Now()

	type args struct {
		m reader.Membership
	}
	tests := []struct {
		name   string
		fields Order
		args   args
		want   bool
	}{
		{
			name:   "Unconfirmed order",
			fields: NewMockOrderBuilder("").Build(),
			args: args{
				m: reader.Membership{},
			},
			want: false,
		},
		{
			name: "Confirmed but out of sync",
			fields: NewMockOrderBuilder("").
				WithConfirmed().
				Build(),
			args: args{
				m: reader.Membership{},
			},
			want: false,
		},
		{
			name: "Confirmed and synced",
			fields: NewMockOrderBuilder("").
				WithConfirmed().
				WithStartTime(now).
				Build(),
			args: args{
				m: reader.Membership{
					ExpireDate: chrono.DateFrom(now.AddDate(1, 0, 1)),
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := tt.fields

			t.Logf("End date: %s", o.EndDate)
			t.Logf("Expire date: %s", tt.args.m.ExpireDate)

			if got := o.IsExpireDateSynced(tt.args.m); got != tt.want {
				t.Errorf("IsExpireDateSynced() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_calibrateOrderKind(t *testing.T) {
	userID := uuid.New().String()

	type args struct {
		m reader.Membership
		e price.Edition
	}
	tests := []struct {
		name string
		args args
		want enum.OrderKind
	}{
		{
			name: "Expired membership",
			args: args{
				m: reader.NewMockMemberBuilder(userID).WithExpiration(time.Now().AddDate(0, 0, -1)).Build(),
				e: price.StdYearEdition,
			},
			want: enum.OrderKindCreate,
		},
		{
			name: "Renewal",
			args: args{
				m: reader.NewMockMemberBuilder(userID).
					Build(),
				e: price.StdYearEdition,
			},
			want: enum.OrderKindRenew,
		},
		{
			name: "Upgrade",
			args: args{
				m: reader.NewMockMemberBuilder(userID).
					Build(),
				e: price.PremiumEdition,
			},
			want: enum.OrderKindUpgrade,
		},
		{
			name: "Standard add-on for premium",
			args: args{
				m: reader.NewMockMemberBuilder(userID).
					WithPrice(faker.PricePrm.Price).
					Build(),
				e: price.StdYearEdition,
			},
			want: enum.OrderKindAddOn,
		},
		{
			name: "Stripe add-on",
			args: args{
				m: reader.NewMockMemberBuilder(userID).
					WithPayMethod(enum.PayMethodStripe).
					Build(),
				e: price.StdYearEdition,
			},
			want: enum.OrderKindAddOn,
		},
		{
			name: "IAP add-on",
			args: args{
				m: reader.NewMockMemberBuilder(userID).
					WithPayMethod(enum.PayMethodApple).
					Build(),
				e: price.StdYearEdition,
			},
			want: enum.OrderKindAddOn,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calibrateOrderKind(tt.args.m, tt.args.e); got != tt.want {
				t.Errorf("calibrateOrderKind() = %v, want %v", got, tt.want)
			}
		})
	}
}
