package invoice

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/pkg/addon"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"testing"
)

func TestConsumeAddOns(t *testing.T) {
	userID := uuid.New().String()

	type args struct {
		addOns []Invoice
	}
	tests := []struct {
		name string
		args args
		want []Invoice
	}{
		{
			name: "Consume add on",
			args: args{
				addOns: []Invoice{
					{
						ID: db.InvoiceID(),
						Edition: price.Edition{
							Tier:  enum.TierStandard,
							Cycle: enum.CycleYear,
						},
						YearMonthDay: dt.YearMonthDay{
							Years:  0,
							Months: 0,
							Days:   48,
						},
						OrderKind:      enum.OrderKindAddOn,
						AddOnSource:    addon.SourceUpgradeCarryOver,
						PaymentMethod:  enum.PayMethodAli,
						CompoundID:     userID,
						OrderID:        null.StringFrom(db.MustOrderID()),
						PriceID:        null.StringFrom("fake-price=-id"),
						CreatedUTC:     chrono.TimeNow(),
						ConsumedUTC:    chrono.Time{},
						DateTimePeriod: dt.DateTimePeriod{},
						CarriedOverUtc: chrono.Time{},
					},
					{
						ID: db.InvoiceID(),
						Edition: price.Edition{
							Tier:  enum.TierStandard,
							Cycle: enum.CycleYear,
						},
						YearMonthDay: dt.YearMonthDay{
							Years:  0,
							Months: 0,
							Days:   15,
						},
						OrderKind:      enum.OrderKindAddOn,
						AddOnSource:    addon.SourceOneTimeToSubCarryOver,
						PaymentMethod:  enum.PayMethodAli,
						CompoundID:     userID,
						OrderID:        null.StringFrom(db.MustOrderID()),
						PriceID:        null.StringFrom("fake-price=-id"),
						CreatedUTC:     chrono.TimeNow(),
						ConsumedUTC:    chrono.Time{},
						DateTimePeriod: dt.DateTimePeriod{},
						CarriedOverUtc: chrono.Time{},
					},
					{
						ID: db.InvoiceID(),
						Edition: price.Edition{
							Tier:  enum.TierStandard,
							Cycle: enum.CycleYear,
						},
						YearMonthDay: dt.YearMonthDay{
							Years:  0,
							Months: 0,
							Days:   30,
						},
						OrderKind:      enum.OrderKindAddOn,
						AddOnSource:    addon.SourceCompensation,
						PaymentMethod:  enum.PayMethodAli,
						CompoundID:     userID,
						OrderID:        null.StringFrom(db.MustOrderID()),
						PriceID:        null.StringFrom("fake-price=-id"),
						CreatedUTC:     chrono.TimeNow(),
						ConsumedUTC:    chrono.Time{},
						DateTimePeriod: dt.DateTimePeriod{},
						CarriedOverUtc: chrono.Time{},
					},
					{
						ID: db.InvoiceID(),
						Edition: price.Edition{
							Tier:  enum.TierStandard,
							Cycle: enum.CycleYear,
						},
						YearMonthDay: dt.YearMonthDay{
							Years:  1,
							Months: 0,
							Days:   1,
						},
						OrderKind:      enum.OrderKindAddOn,
						AddOnSource:    "",
						PaymentMethod:  enum.PayMethodAli,
						CompoundID:     userID,
						OrderID:        null.StringFrom(db.MustOrderID()),
						PriceID:        null.StringFrom("fake-price=-id"),
						CreatedUTC:     chrono.TimeNow(),
						ConsumedUTC:    chrono.Time{},
						DateTimePeriod: dt.DateTimePeriod{},
						CarriedOverUtc: chrono.Time{},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConsumeAddOns(tt.args.addOns)

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}
