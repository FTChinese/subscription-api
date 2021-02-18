package subs

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/collection"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/guregu/null"
	"reflect"
	"testing"
	"time"
)

func TestNewAddOn(t *testing.T) {

	order := MockOrder(faker.PriceStdYear, enum.OrderKindAddOn)

	type args struct {
		o Order
	}
	tests := []struct {
		name string
		args args
		want AddOn
	}{
		{
			name: "New add on from an order",
			args: args{o: order},
			want: AddOn{
				ID:                 "",
				Edition:            faker.PriceStdYear.Original.Edition,
				CycleCount:         1,
				DaysRemained:       1,
				IsUpgradeCarryOver: false,
				PaymentMethod:      enum.PayMethodAli,
				CompoundID:         order.CompoundID,
				OrderID:            null.StringFrom(order.ID),
				PlanID:             null.StringFrom(order.PlanID),
				CreatedUTC:         chrono.TimeNow(),
				ConsumedUTC:        chrono.Time{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewAddOn(tt.args.o)
			tt.want.ID = got.ID
			tt.want.CreatedUTC = got.CreatedUTC

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewAddOn() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewUpgradeCarryOver(t *testing.T) {
	order := MockOrder(faker.PricePrm, enum.OrderKindUpgrade)

	type args struct {
		o Order
		m reader.Membership
	}
	tests := []struct {
		name string
		args args
		want AddOn
	}{
		{
			name: "Upgrade carry over",
			args: args{
				o: order,
				m: reader.Membership{
					Edition:       faker.PriceStdYear.Original.Edition,
					ExpireDate:    chrono.DateFrom(time.Now().AddDate(0, 0, 10)),
					PaymentMethod: enum.PayMethodWx,
					FtcPlanID:     null.StringFrom(faker.PriceStdYear.Original.ID),
				},
			},
			want: AddOn{
				ID:                 "",
				Edition:            faker.PriceStdYear.Original.Edition,
				CycleCount:         0,
				DaysRemained:       10,
				IsUpgradeCarryOver: true,
				PaymentMethod:      enum.PayMethodWx,
				CompoundID:         order.CompoundID,
				OrderID:            null.StringFrom(order.ID),
				PlanID:             null.StringFrom(faker.PriceStdYear.Original.ID),
				CreatedUTC:         chrono.TimeNow(),
				ConsumedUTC:        chrono.Time{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewUpgradeCarryOver(tt.args.o, tt.args.m)
			tt.want.ID = got.ID
			tt.want.CreatedUTC = got.CreatedUTC

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewUpgradeCarryOver() = %v\n, want %v", got, tt.want)
			}
		})
	}
}

func TestAddOn_IsZero(t *testing.T) {

	tests := []struct {
		name   string
		fields AddOn
		want   bool
	}{
		{
			name: "Is zero",
			fields: AddOn{
				ID: "",
			},
			want: true,
		},
		{
			name: "Not zero",
			fields: AddOn{
				ID: db.AddOnID(),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := tt.fields
			if got := a.IsZero(); got != tt.want {
				t.Errorf("IsZero() = %v, want %v", got, tt.want)
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
			fields: NewAddOn(MockOrder(faker.PriceStdYear, enum.OrderKindCreate)),
			want:   367,
		},
		{
			name:   "Days of a monthly cycle",
			fields: NewAddOn(MockOrder(faker.PriceStdMonth, enum.OrderKindCreate)),
			want:   32,
		},
		{
			name: "Remaining days for upgrade",
			fields: NewUpgradeCarryOver(MockOrder(faker.PricePrm, enum.OrderKindUpgrade), reader.Membership{
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
			fields: NewAddOn(MockOrder(faker.PriceStdYear, enum.OrderKindAddOn)),
			want: reader.ReservedDays{
				Standard: 367,
				Premium:  0,
			},
		},
		{
			name:   "Reserved days for premium",
			fields: NewAddOn(MockOrder(faker.PricePrm, enum.OrderKindAddOn)),
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

func TestConsumeAddOns(t *testing.T) {

	stripeM := reader.NewMockMemberBuilder("").
		WithPayMethod(enum.PayMethodStripe).
		WithExpiration(time.Now()).
		Build()

	addOn := NewMockAddOnBuilder().
		WithUserIDs(stripeM.MemberID).
		BuildNew()

	type args struct {
		addOns []AddOn
		m      reader.Membership
	}
	tests := []struct {
		name string
		args args
		want AddOnConsumed
	}{
		{
			name: "Transfer add-on",
			args: args{
				addOns: []AddOn{
					addOn,
				},
				m: stripeM,
			},
			want: AddOnConsumed{
				AddOnIDs: collection.StringSet{
					addOn.ID: nil,
				},
				Membership: reader.Membership{
					MemberID:      stripeM.MemberID,
					Edition:       addOn.Edition,
					LegacyTier:    null.Int{},
					LegacyExpire:  null.Int{},
					ExpireDate:    chrono.DateFrom(time.Now().AddDate(1, 0, 1)),
					PaymentMethod: addOn.PaymentMethod,
					FtcPlanID:     addOn.PlanID,
				}.Sync(),
				Snapshot: stripeM.Snapshot(reader.FtcArchiver(enum.OrderKindAddOn)),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConsumeAddOns(tt.args.addOns, tt.args.m)
			tt.want.Snapshot.SnapshotID = got.Snapshot.SnapshotID
			tt.want.Snapshot.CreatedUTC = got.Snapshot.CreatedUTC

			if !reflect.DeepEqual(got.AddOnIDs, tt.want.AddOnIDs) {
				t.Errorf("ConsumeAddOns().AddOnIDs = %v\n, want %v", got.AddOnIDs, tt.want.AddOnIDs)
			}

			if !reflect.DeepEqual(got.Membership, tt.want.Membership) {
				t.Errorf("ConsumeAddOns().Membership = %v\n, want %v", got.Membership, tt.want.Membership)
			}

			if !reflect.DeepEqual(got.Snapshot, tt.want.Snapshot) {
				t.Errorf("ConsumeAddOns().Snapshot = %v\n, want %v", got.Snapshot, tt.want.Snapshot)
			}
		})
	}
}

func TestGroupAddOns(t *testing.T) {
	mocker := NewMockAddOnBuilder()

	a1 := mocker.BuildNew()
	a2 := mocker.WithPlan(faker.PricePrm).BuildNew()
	a3 := mocker.BuildNew()

	type args struct {
		addOns []AddOn
	}
	tests := []struct {
		name string
		args args
		want map[enum.Tier][]AddOn
	}{
		{
			name: "Group add-ons",
			args: args{
				addOns: []AddOn{
					a1,
					a2,
					a3,
				},
			},
			want: map[enum.Tier][]AddOn{
				enum.TierStandard: {
					a1,
					a3,
				},
				enum.TierPremium: {
					a2,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GroupAddOns(tt.args.addOns)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GroupAddOns() = %v, want %v", got, tt.want)
			}

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}

func TestReduceAddOns(t *testing.T) {
	mocker := NewMockAddOnBuilder()
	a1 := mocker.BuildNew()
	a2 := mocker.BuildNew()
	a3 := mocker.BuildUpgrade()
	a4 := mocker.WithPlan(faker.PriceStdMonth).BuildNew()

	type args struct {
		addOns []AddOn
	}
	tests := []struct {
		name string
		args args
		want AddOnSum
	}{
		{
			name: "Reduce addons",
			args: args{
				addOns: []AddOn{
					a1,
					a2,
					a3,
					a4,
				},
			},
			want: AddOnSum{
				Years:  2,
				Months: 1,
				Days:   3 + int(a3.DaysRemained),
				Latest: a1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ReduceAddOns(tt.args.addOns)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReduceAddOns() = %v\n, want %v", got, tt.want)
				return
			}

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}

func TestAddOnSum_Membership(t *testing.T) {
	stripeM := reader.NewMockMemberBuilder("").
		WithPayMethod(enum.PayMethodStripe).
		WithExpiration(time.Now()).
		Build()

	a := NewMockAddOnBuilder().
		WithUserIDs(stripeM.MemberID).
		BuildNew()

	type fields struct {
		Years  int
		Months int
		Days   int
		Latest AddOn
	}
	type args struct {
		m reader.Membership
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   reader.Membership
	}{
		{
			name: "Update membership from add-on",
			fields: fields{
				Years:  2,
				Months: 1,
				Days:   6,
				Latest: a,
			},
			args: args{
				m: stripeM,
			},
			want: reader.Membership{
				MemberID:      stripeM.MemberID,
				Edition:       a.Edition,
				LegacyTier:    null.Int{},
				LegacyExpire:  null.Int{},
				ExpireDate:    chrono.DateFrom(stripeM.ExpireDate.AddDate(2, 1, 6)),
				PaymentMethod: a.PaymentMethod,
				FtcPlanID:     a.PlanID,
				ReservedDays:  reader.ReservedDays{},
			}.Sync(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := AddOnSum{
				Years:  tt.fields.Years,
				Months: tt.fields.Months,
				Days:   tt.fields.Days,
				Latest: tt.fields.Latest,
			}
			if got := s.Membership(tt.args.m); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Membership() = %v\n, want %v", got, tt.want)
			}
		})
	}
}
