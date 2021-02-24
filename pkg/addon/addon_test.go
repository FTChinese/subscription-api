package addon

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"reflect"
	"testing"
)

func TestAddOn_IsZero(t *testing.T) {
	type fields struct {
		ID              string
		Edition         price.Edition
		CycleCount      int64
		DaysRemained    int64
		CarryOverSource CarryOverSource
		PaymentMethod   enum.PayMethod
		CompoundID      string
		OrderID         null.String
		PlanID          null.String
		CreatedUTC      chrono.Time
		ConsumedUTC     chrono.Time
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name:   "Is zero",
			fields: fields{},
			want:   true,
		},
		{
			name: "Not zero",
			fields: fields{
				ID:              db.AddOnID(),
				Edition:         price.StdYearEdition,
				CycleCount:      1,
				DaysRemained:    1,
				CarryOverSource: CarryOverFromUpgrade,
				PaymentMethod:   enum.PayMethodAli,
				CompoundID:      uuid.New().String(),
				OrderID:         null.StringFrom(db.MustOrderID()),
				PlanID:          null.StringFrom(faker.PriceStdYear.Original.ID),
				CreatedUTC:      chrono.TimeNow(),
				ConsumedUTC:     chrono.Time{},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := AddOn{
				ID:              tt.fields.ID,
				Edition:         tt.fields.Edition,
				CycleCount:      tt.fields.CycleCount,
				DaysRemained:    tt.fields.DaysRemained,
				CarryOverSource: tt.fields.CarryOverSource,
				PaymentMethod:   tt.fields.PaymentMethod,
				CompoundID:      tt.fields.CompoundID,
				OrderID:         tt.fields.OrderID,
				PlanID:          tt.fields.PlanID,
				CreatedUTC:      tt.fields.CreatedUTC,
				ConsumedUTC:     tt.fields.ConsumedUTC,
			}
			if got := a.IsZero(); got != tt.want {
				t.Errorf("IsZero() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddOn_getDays(t *testing.T) {
	type fields struct {
		ID              string
		Edition         price.Edition
		CycleCount      int64
		DaysRemained    int64
		CarryOverSource CarryOverSource
		PaymentMethod   enum.PayMethod
		CompoundID      string
		OrderID         null.String
		PlanID          null.String
		CreatedUTC      chrono.Time
		ConsumedUTC     chrono.Time
	}
	tests := []struct {
		name   string
		fields fields
		want   int64
	}{
		{
			name: "Get days of a year",
			fields: fields{
				ID:              db.AddOnID(),
				Edition:         price.StdYearEdition,
				CycleCount:      1,
				DaysRemained:    1,
				CarryOverSource: "",
				PaymentMethod:   enum.PayMethodAli,
				CompoundID:      uuid.New().String(),
				OrderID:         null.StringFrom(db.MustOrderID()),
				PlanID:          null.StringFrom(faker.PriceStdYear.Original.ID),
				CreatedUTC:      chrono.TimeNow(),
				ConsumedUTC:     chrono.Time{},
			},
			want: 367,
		},
		{
			name: "Get days of a month",
			fields: fields{
				ID:              db.AddOnID(),
				Edition:         price.StdMonthEdition,
				CycleCount:      1,
				DaysRemained:    1,
				CarryOverSource: "",
				PaymentMethod:   enum.PayMethodAli,
				CompoundID:      uuid.New().String(),
				OrderID:         null.StringFrom(db.MustOrderID()),
				PlanID:          null.StringFrom(faker.PriceStdMonth.Original.ID),
				CreatedUTC:      chrono.TimeNow(),
				ConsumedUTC:     chrono.Time{},
			},
			want: 32,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := AddOn{
				ID:              tt.fields.ID,
				Edition:         tt.fields.Edition,
				CycleCount:      tt.fields.CycleCount,
				DaysRemained:    tt.fields.DaysRemained,
				CarryOverSource: tt.fields.CarryOverSource,
				PaymentMethod:   tt.fields.PaymentMethod,
				CompoundID:      tt.fields.CompoundID,
				OrderID:         tt.fields.OrderID,
				PlanID:          tt.fields.PlanID,
				CreatedUTC:      tt.fields.CreatedUTC,
				ConsumedUTC:     tt.fields.ConsumedUTC,
			}
			if got := a.GetDays(); got != tt.want {
				t.Errorf("GetDays() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddOn_ToReservedDays(t *testing.T) {
	type fields struct {
		ID              string
		Edition         price.Edition
		CycleCount      int64
		DaysRemained    int64
		CarryOverSource CarryOverSource
		PaymentMethod   enum.PayMethod
		CompoundID      string
		OrderID         null.String
		PlanID          null.String
		CreatedUTC      chrono.Time
		ConsumedUTC     chrono.Time
	}
	tests := []struct {
		name   string
		fields fields
		want   ReservedDays
	}{
		{
			name: "Reserved days of standard year",
			fields: fields{
				ID:              db.AddOnID(),
				Edition:         price.StdYearEdition,
				CycleCount:      1,
				DaysRemained:    1,
				CarryOverSource: "",
				PaymentMethod:   enum.PayMethodAli,
				CompoundID:      uuid.New().String(),
				OrderID:         null.StringFrom(db.MustOrderID()),
				PlanID:          null.StringFrom(faker.PriceStdYear.Original.ID),
				CreatedUTC:      chrono.TimeNow(),
				ConsumedUTC:     chrono.Time{},
			},
			want: ReservedDays{
				Standard: 367,
				Premium:  0,
			},
		},
		{
			name: "Reserved days of premium year",
			fields: fields{
				ID:              db.AddOnID(),
				Edition:         price.PremiumEdition,
				CycleCount:      1,
				DaysRemained:    1,
				CarryOverSource: "",
				PaymentMethod:   enum.PayMethodAli,
				CompoundID:      uuid.New().String(),
				OrderID:         null.StringFrom(db.MustOrderID()),
				PlanID:          null.StringFrom(faker.PricePrm.Original.ID),
				CreatedUTC:      chrono.TimeNow(),
				ConsumedUTC:     chrono.Time{},
			},
			want: ReservedDays{
				Standard: 0,
				Premium:  367,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := AddOn{
				ID:              tt.fields.ID,
				Edition:         tt.fields.Edition,
				CycleCount:      tt.fields.CycleCount,
				DaysRemained:    tt.fields.DaysRemained,
				CarryOverSource: tt.fields.CarryOverSource,
				PaymentMethod:   tt.fields.PaymentMethod,
				CompoundID:      tt.fields.CompoundID,
				OrderID:         tt.fields.OrderID,
				PlanID:          tt.fields.PlanID,
				CreatedUTC:      tt.fields.CreatedUTC,
				ConsumedUTC:     tt.fields.ConsumedUTC,
			}
			if got := a.ToReservedDays(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToReservedDays() = %v, want %v", got, tt.want)
			}
		})
	}
}
