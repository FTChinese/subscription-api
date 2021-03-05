package addon

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/lib/collection"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"reflect"
	"testing"
)

func Test_group(t *testing.T) {
	userID := uuid.New().String()
	addOnStdMonth := AddOn{
		ID:              db.AddOnID(),
		Edition:         price.StdMonthEdition,
		CycleCount:      1,
		DaysRemained:    1,
		CarryOverSource: "",
		PaymentMethod:   enum.PayMethodAli,
		CompoundID:      userID,
		OrderID:         null.StringFrom(db.MustOrderID()),
		PlanID:          null.StringFrom(faker.PriceStdMonth.ID),
		CreatedUTC:      chrono.TimeNow(),
		ConsumedUTC:     chrono.Time{},
	}

	addOnStdYear := AddOn{
		ID:              db.AddOnID(),
		Edition:         price.StdYearEdition,
		CycleCount:      1,
		DaysRemained:    1,
		CarryOverSource: "",
		PaymentMethod:   enum.PayMethodWx,
		CompoundID:      userID,
		OrderID:         null.StringFrom(db.MustOrderID()),
		PlanID:          null.StringFrom(faker.PriceStdYear.ID),
		CreatedUTC:      chrono.TimeNow(),
		ConsumedUTC:     chrono.Time{},
	}

	addOnPrm := AddOn{
		ID:              db.AddOnID(),
		Edition:         price.PremiumEdition,
		CycleCount:      1,
		DaysRemained:    1,
		CarryOverSource: "",
		PaymentMethod:   enum.PayMethodAli,
		CompoundID:      userID,
		OrderID:         null.StringFrom(db.MustOrderID()),
		PlanID:          null.StringFrom(faker.PricePrm.ID),
		CreatedUTC:      chrono.TimeNow(),
		ConsumedUTC:     chrono.Time{},
	}

	addOnUpgradeCarryOver := AddOn{
		ID:              db.AddOnID(),
		Edition:         price.StdYearEdition,
		CycleCount:      0,
		DaysRemained:    30,
		CarryOverSource: SourceUpgradeCarryOver,
		PaymentMethod:   enum.PayMethodAli,
		CompoundID:      userID,
		OrderID:         null.String{},
		PlanID:          null.StringFrom(faker.PriceStdYear.ID),
		CreatedUTC:      chrono.TimeNow(),
		ConsumedUTC:     chrono.Time{},
	}

	addOnStripeCarryOver := AddOn{
		ID:              db.AddOnID(),
		Edition:         price.StdYearEdition,
		CycleCount:      0,
		DaysRemained:    30,
		CarryOverSource: SourceOneTimeToSubCarryOver,
		PaymentMethod:   enum.PayMethodAli,
		CompoundID:      userID,
		OrderID:         null.String{},
		PlanID:          null.StringFrom(faker.PriceStdYear.ID),
		CreatedUTC:      chrono.TimeNow(),
		ConsumedUTC:     chrono.Time{},
	}

	type args struct {
		addOns []AddOn
	}
	tests := []struct {
		name string
		args args
		want map[enum.Tier][]AddOn
	}{
		{
			name: "Group addons",
			args: args{
				addOns: []AddOn{
					addOnStdMonth,
					addOnStdYear,
					addOnPrm,
					addOnUpgradeCarryOver,
					addOnStripeCarryOver,
				},
			},
			want: map[enum.Tier][]AddOn{
				enum.TierStandard: {
					addOnStdMonth,
					addOnStdYear,
					addOnUpgradeCarryOver,
					addOnStripeCarryOver,
				},
				enum.TierPremium: {
					addOnPrm,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := group(tt.args.addOns); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("group() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_reduce(t *testing.T) {
	userID := uuid.New().String()
	addOnStdMonth := AddOn{
		ID:              db.AddOnID(),
		Edition:         price.StdMonthEdition,
		CycleCount:      1,
		DaysRemained:    1,
		CarryOverSource: "",
		PaymentMethod:   enum.PayMethodAli,
		CompoundID:      userID,
		OrderID:         null.StringFrom(db.MustOrderID()),
		PlanID:          null.StringFrom(faker.PriceStdMonth.ID),
		CreatedUTC:      chrono.TimeNow(),
		ConsumedUTC:     chrono.Time{},
	}

	addOnStdYear := AddOn{
		ID:              db.AddOnID(),
		Edition:         price.StdYearEdition,
		CycleCount:      1,
		DaysRemained:    1,
		CarryOverSource: "",
		PaymentMethod:   enum.PayMethodWx,
		CompoundID:      userID,
		OrderID:         null.StringFrom(db.MustOrderID()),
		PlanID:          null.StringFrom(faker.PriceStdYear.ID),
		CreatedUTC:      chrono.TimeNow(),
		ConsumedUTC:     chrono.Time{},
	}

	type args struct {
		addOns []AddOn
	}
	tests := []struct {
		name string
		args args
		want Sum
	}{
		{
			name: "Add all standard add-ons",
			args: args{
				addOns: []AddOn{
					addOnStdMonth,
					addOnStdYear,
				},
			},
			want: Sum{
				IDs: collection.StringSet{
					addOnStdMonth.ID: nil,
					addOnStdYear.ID:  nil,
				},
				Years:  1,
				Months: 1,
				Days:   2,
				Latest: addOnStdMonth,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := reduce(tt.args.addOns); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("reduce() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGroupAndReduce(t *testing.T) {
	userID := uuid.New().String()
	addOnStdMonth := AddOn{
		ID:              db.AddOnID(),
		Edition:         price.StdMonthEdition,
		CycleCount:      1,
		DaysRemained:    1,
		CarryOverSource: "",
		PaymentMethod:   enum.PayMethodAli,
		CompoundID:      userID,
		OrderID:         null.StringFrom(db.MustOrderID()),
		PlanID:          null.StringFrom(faker.PriceStdMonth.ID),
		CreatedUTC:      chrono.TimeNow(),
		ConsumedUTC:     chrono.Time{},
	}

	addOnStdYear := AddOn{
		ID:              db.AddOnID(),
		Edition:         price.StdYearEdition,
		CycleCount:      1,
		DaysRemained:    1,
		CarryOverSource: "",
		PaymentMethod:   enum.PayMethodWx,
		CompoundID:      userID,
		OrderID:         null.StringFrom(db.MustOrderID()),
		PlanID:          null.StringFrom(faker.PriceStdYear.ID),
		CreatedUTC:      chrono.TimeNow(),
		ConsumedUTC:     chrono.Time{},
	}

	addOnPrm := AddOn{
		ID:              db.AddOnID(),
		Edition:         price.PremiumEdition,
		CycleCount:      1,
		DaysRemained:    1,
		CarryOverSource: "",
		PaymentMethod:   enum.PayMethodAli,
		CompoundID:      userID,
		OrderID:         null.StringFrom(db.MustOrderID()),
		PlanID:          null.StringFrom(faker.PricePrm.ID),
		CreatedUTC:      chrono.TimeNow(),
		ConsumedUTC:     chrono.Time{},
	}

	addOnUpgradeCarryOver := AddOn{
		ID:              db.AddOnID(),
		Edition:         price.StdYearEdition,
		CycleCount:      0,
		DaysRemained:    30,
		CarryOverSource: SourceUpgradeCarryOver,
		PaymentMethod:   enum.PayMethodAli,
		CompoundID:      userID,
		OrderID:         null.String{},
		PlanID:          null.StringFrom(faker.PriceStdYear.ID),
		CreatedUTC:      chrono.TimeNow(),
		ConsumedUTC:     chrono.Time{},
	}

	addOnStripeCarryOver := AddOn{
		ID:              db.AddOnID(),
		Edition:         price.StdYearEdition,
		CycleCount:      0,
		DaysRemained:    30,
		CarryOverSource: SourceOneTimeToSubCarryOver,
		PaymentMethod:   enum.PayMethodAli,
		CompoundID:      userID,
		OrderID:         null.String{},
		PlanID:          null.StringFrom(faker.PriceStdYear.ID),
		CreatedUTC:      chrono.TimeNow(),
		ConsumedUTC:     chrono.Time{},
	}

	type args struct {
		addOns []AddOn
	}
	tests := []struct {
		name string
		args args
		want map[enum.Tier]Sum
	}{
		{
			name: "Group and reduce addons",
			args: args{
				addOns: []AddOn{
					addOnStdMonth,
					addOnStdYear,
					addOnPrm,
					addOnUpgradeCarryOver,
					addOnStripeCarryOver,
				},
			},
			want: map[enum.Tier]Sum{
				enum.TierStandard: {
					IDs: collection.StringSet{
						addOnStdMonth.ID:         nil,
						addOnStdYear.ID:          nil,
						addOnUpgradeCarryOver.ID: nil,
						addOnStripeCarryOver.ID:  nil,
					},
					Years:  1,
					Months: 1,
					Days:   1 + 1 + 30 + 30,
					Latest: addOnStdMonth,
				},
				enum.TierPremium: {
					IDs: collection.StringSet{
						addOnPrm.ID: nil,
					},
					Years:  1,
					Months: 0,
					Days:   1,
					Latest: addOnPrm,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GroupAndReduce(tt.args.addOns); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GroupAndReduce() = %v, want %v", got, tt.want)
			}
		})
	}
}
