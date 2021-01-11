package subs

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/dt"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"reflect"
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
			fields: MockOrder(faker.PlanStdYear, enum.OrderKindCreate),
			args: args{
				m: reader.Membership{},
			},
			want: false,
		},
		{
			name:   "Confirmed but out of sync",
			fields: MockOrder(faker.PlanStdYear, enum.OrderKindCreate).newOrRenewalConfirm(chrono.TimeFrom(now), chrono.Date{}),
			args: args{
				m: reader.Membership{},
			},
			want: false,
		},
		{
			name:   "Confirmed and synced",
			fields: MockOrder(faker.PlanStdYear, enum.OrderKindCreate).newOrRenewalConfirm(chrono.TimeFrom(now), chrono.Date{}),
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

			if got := o.IsSynced(tt.args.m); got != tt.want {
				t.Errorf("IsSynced() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOrder_newOrRenewalConfirm(t *testing.T) {
	now := time.Now()

	exp := now.AddDate(0, 1, 0)

	type args struct {
		confirmedAt chrono.Time
		exp         chrono.Date
	}
	tests := []struct {
		name   string
		fields Order
		args   args
		want   dt.DateRange
	}{
		{
			name:   "Order for new subscription",
			fields: MockOrder(faker.PlanStdYear, enum.OrderKindCreate),
			args: args{
				confirmedAt: chrono.TimeFrom(now),
				exp:         chrono.Date{},
			},
			want: dt.DateRange{
				StartDate: chrono.DateFrom(now),
				EndDate:   chrono.DateFrom(now.AddDate(1, 0, 1)),
			},
		},
		{
			name:   "Order for renewal subscription",
			fields: MockOrder(faker.PlanStdYear, enum.OrderKindRenew),
			args: args{
				confirmedAt: chrono.TimeFrom(now),
				exp:         chrono.DateFrom(exp),
			},
			want: dt.DateRange{
				StartDate: chrono.DateFrom(exp),
				EndDate:   chrono.DateFrom(exp.AddDate(1, 0, 1)),
			},
		},
		{
			name:   "Order for addon",
			fields: MockOrder(faker.PlanStdYear, enum.OrderKindAddOn),
			args: args{
				confirmedAt: chrono.TimeFrom(now),
				exp:         chrono.Date{},
			},
			want: dt.DateRange{
				StartDate: chrono.DateFrom(now),
				EndDate:   chrono.DateFrom(now.AddDate(1, 0, 1)),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := tt.fields
			if got := o.newOrRenewalConfirm(tt.args.confirmedAt, tt.args.exp); !reflect.DeepEqual(got.DateRange, tt.want) {
				t.Errorf("newOrRenewalConfirm() = %v, want %v", got.DateRange, tt.want)
			}
		})
	}
}

func TestOrder_upgradeConfirm(t *testing.T) {
	now := time.Now()
	type args struct {
		confirmedAt chrono.Time
	}
	tests := []struct {
		name   string
		fields Order
		args   args
		want   dt.DateRange
	}{
		{
			name:   "Confirm upgrading order",
			fields: MockOrder(faker.PlanStdYear, enum.OrderKindUpgrade),
			args: args{
				confirmedAt: chrono.TimeFrom(now),
			},
			want: dt.DateRange{
				StartDate: chrono.DateFrom(now),
				EndDate:   chrono.DateFrom(now.AddDate(1, 0, 1)),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := tt.fields
			if got := o.upgradeConfirm(tt.args.confirmedAt); !reflect.DeepEqual(got.DateRange, tt.want) {
				t.Errorf("upgradeConfirm() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddOn_GetDays(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name   string
		fields AddOn
		want   int64
	}{
		{
			name:   "Days of a yearly cycle",
			fields: NewAddOn(MockOrder(faker.PlanStdYear, enum.OrderKindCreate)),
			want:   367,
		},
		{
			name:   "Days of a monthly cycle",
			fields: NewAddOn(MockOrder(faker.PlanStdMonth, enum.OrderKindCreate)),
			want:   32,
		},
		{
			name: "Remaining days for upgrade",
			fields: NewUpgradeAddOn(MockOrder(faker.PlanPrm, enum.OrderKindUpgrade), reader.Membership{
				ExpireDate: chrono.DateFrom(now.AddDate(0, 0, 10)),
			}),
			want: 10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := tt.fields
			if got := a.GetDays(); got != tt.want {
				t.Errorf("GetDays() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddOn_ToReservedDays(t *testing.T) {

	tests := []struct {
		name   string
		fields AddOn
		want   reader.ReservedDays
	}{
		{
			name:   "Reserved days for standard year",
			fields: NewAddOn(MockOrder(faker.PlanStdYear, enum.OrderKindAddOn)),
			want: reader.ReservedDays{
				Standard: 367,
				Premium:  0,
			},
		},
		{
			name:   "Reserved days for premium",
			fields: NewAddOn(MockOrder(faker.PlanPrm, enum.OrderKindAddOn)),
			want: reader.ReservedDays{
				Standard: 0,
				Premium:  367,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := AddOn{
				ID:            tt.fields.ID,
				Edition:       tt.fields.Edition,
				CycleCount:    tt.fields.CycleCount,
				DaysRemained:  tt.fields.DaysRemained,
				PaymentMethod: tt.fields.PaymentMethod,
				OrderID:       tt.fields.OrderID,
				CompoundID:    tt.fields.CompoundID,
				CreatedUTC:    tt.fields.CreatedUTC,
				ConsumedUTC:   tt.fields.ConsumedUTC,
			}
			if got := a.ToReservedDays(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToReservedDays() = %v, want %v", got, tt.want)
			}
		})
	}
}
