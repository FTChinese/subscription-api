package subs

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/collection"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/product"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
)

var ftcID = uuid.New().String()

var stdAddOn = AddOn{
	ID:                 ftcID,
	Edition:            faker.PlanStdYear.Edition,
	CycleCount:         1,
	DaysRemained:       1,
	IsUpgradeCarryOver: false,
	PaymentMethod:      enum.PayMethodWx,
	CompoundID:         ftcID,
	OrderID:            null.StringFrom(db.MustOrderID()),
	PlanID:             null.StringFrom(faker.PlanStdYear.ID),
	CreatedUTC:         chrono.TimeNow(),
	ConsumedUTC:        chrono.Time{},
}

var prmAddOn = AddOn{
	ID:                 ftcID,
	Edition:            faker.PlanPrm.Edition,
	CycleCount:         1,
	DaysRemained:       1,
	IsUpgradeCarryOver: false,
	PaymentMethod:      enum.PayMethodAli,
	CompoundID:         ftcID,
	OrderID:            null.StringFrom(db.MustOrderID()),
	PlanID:             null.StringFrom(faker.PlanPrm.ID),
	CreatedUTC:         chrono.TimeNow(),
	ConsumedUTC:        chrono.Time{},
}

func TestNewAddOn(t *testing.T) {

	order := MockOrder(faker.PlanStdYear, enum.OrderKindAddOn)

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
				Edition:            faker.PlanStdYear.Edition,
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
	order := MockOrder(faker.PlanPrm, enum.OrderKindUpgrade)

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
					Edition:       faker.PlanStdYear.Edition,
					ExpireDate:    chrono.DateFrom(time.Now().AddDate(0, 0, 10)),
					PaymentMethod: enum.PayMethodWx,
					FtcPlanID:     null.StringFrom(faker.PlanStdYear.ID),
				},
			},
			want: AddOn{
				ID:                 "",
				Edition:            faker.PlanStdYear.Edition,
				CycleCount:         0,
				DaysRemained:       10,
				IsUpgradeCarryOver: true,
				PaymentMethod:      enum.PayMethodWx,
				CompoundID:         order.CompoundID,
				OrderID:            null.StringFrom(order.ID),
				PlanID:             null.StringFrom(faker.PlanStdYear.ID),
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
			fields: NewUpgradeCarryOver(MockOrder(faker.PlanPrm, enum.OrderKindUpgrade), reader.Membership{
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

func Test_groupAddOns(t *testing.T) {

	type args struct {
		l []AddOn
	}
	tests := []struct {
		name string
		args args
		want map[product.Edition][]AddOn
	}{
		{
			name: "Group add-ons",
			args: args{
				l: []AddOn{
					prmAddOn,
					stdAddOn,
				},
			},
			want: map[product.Edition][]AddOn{
				product.StdYearEdition: {
					stdAddOn,
				},
				product.PremiumEdition: {
					prmAddOn,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := groupAddOns(tt.args.l); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("groupAddOns() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sumAddOns(t *testing.T) {
	type args struct {
		addOns []AddOn
	}
	tests := []struct {
		name string
		args args
		want product.Duration
	}{
		{
			name: "sum add on",
			args: args{
				addOns: []AddOn{
					NewAddOn(MockOrder(faker.PlanStdYear, enum.OrderKindAddOn)),
					NewAddOn(MockOrder(faker.PlanStdYear, enum.OrderKindAddOn)),
					NewAddOn(MockOrder(faker.PlanStdYear, enum.OrderKindAddOn)),
				},
			},
			want: product.Duration{
				CycleCount: 3,
				ExtraDays:  3,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sumAddOns(tt.args.addOns); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sumAddOns() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newMembershipFromAddOn(t *testing.T) {

	ftcID := uuid.New().String()

	stripeM := reader.NewMockMemberBuilder(ftcID).
		WithPayMethod(enum.PayMethodStripe).
		Build()

	prmM := reader.NewMockMemberBuilder(ftcID).
		WithPlan(faker.PlanPrm.Plan).
		WithExpiration(time.Now()).
		Build()

	type args struct {
		addOns []AddOn
		m      reader.Membership
	}
	tests := []struct {
		name string
		args args
		want reader.Membership
	}{
		{
			name: "Add add-on",
			args: args{
				addOns: []AddOn{
					NewAddOn(MockOrder(faker.PlanStdYear, enum.OrderKindAddOn)),
					NewAddOn(MockOrder(faker.PlanStdYear, enum.OrderKindAddOn)),
					NewAddOn(MockOrder(faker.PlanStdYear, enum.OrderKindAddOn)),
				},
				m: stripeM,
			},
			want: reader.Membership{
				MemberID:      stripeM.MemberID,
				Edition:       faker.PlanStdYear.Edition,
				LegacyTier:    null.Int{},
				LegacyExpire:  null.Int{},
				ExpireDate:    chrono.DateFrom(stripeM.ExpireDate.AddDate(3, 0, 3)),
				PaymentMethod: enum.PayMethodAli,
				FtcPlanID:     null.StringFrom(faker.PlanStdYear.ID),
			}.Sync(),
		},
		{
			name: "Restored carry-over of a upgrade",
			args: args{
				addOns: []AddOn{
					{
						ID:                 db.AddOnID(),
						Edition:            faker.PlanStdYear.Edition,
						CycleCount:         0,
						DaysRemained:       30,
						IsUpgradeCarryOver: true,
						PaymentMethod:      enum.PayMethodWx,
						CompoundID:         ftcID,
						OrderID:            null.StringFrom(db.MustOrderID()),
						PlanID:             null.StringFrom(faker.PlanStdYear.ID),
						CreatedUTC:         chrono.TimeNow(),
					},
				},
				m: prmM,
			},
			want: reader.Membership{
				MemberID:      prmM.MemberID,
				Edition:       faker.PlanStdYear.Edition,
				LegacyTier:    null.Int{},
				LegacyExpire:  null.Int{},
				ExpireDate:    chrono.DateFrom(time.Now().AddDate(0, 0, 30)),
				PaymentMethod: enum.PayMethodWx,
				FtcPlanID:     null.StringFrom(faker.PlanStdYear.ID),
			}.Sync(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newMembershipFromAddOn(tt.args.addOns, tt.args.m); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newMembershipFromAddOn() = %v\n, want %v", got, tt.want)
			}
		})
	}
}

func TestTransferAddOn(t *testing.T) {

	stripeM := reader.NewMockMemberBuilder(ftcID).
		WithPayMethod(enum.PayMethodStripe).
		WithExpiration(time.Now()).
		Build()

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
					stdAddOn,
				},
				m: stripeM,
			},
			want: AddOnConsumed{
				AddOnIDs: collection.StringSet{
					stdAddOn.ID: nil,
				},
				Membership: reader.Membership{
					MemberID:      stripeM.MemberID,
					Edition:       faker.PlanStdYear.Edition,
					LegacyTier:    null.Int{},
					LegacyExpire:  null.Int{},
					ExpireDate:    chrono.DateFrom(time.Now().AddDate(1, 0, 1)),
					PaymentMethod: enum.PayMethodWx,
					FtcPlanID:     null.StringFrom(faker.PlanStdYear.ID),
				}.Sync(),
				Snapshot: stripeM.Snapshot(reader.FtcArchiver(enum.OrderKindAddOn)),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TransferAddOn(tt.args.addOns, tt.args.m)
			tt.want.Snapshot.SnapshotID = got.Snapshot.SnapshotID
			tt.want.Snapshot.CreatedUTC = got.Snapshot.CreatedUTC

			if !reflect.DeepEqual(got.AddOnIDs, tt.want.AddOnIDs) {
				t.Errorf("TransferAddOn().AddOnIDs = %v\n, want %v", got.AddOnIDs, tt.want.AddOnIDs)
			}

			if !reflect.DeepEqual(got.Membership, tt.want.Membership) {
				t.Errorf("TransferAddOn().Membership = %v\n, want %v", got.Membership, tt.want.Membership)
			}

			if !reflect.DeepEqual(got.Snapshot, tt.want.Snapshot) {
				t.Errorf("TransferAddOn().Snapshot = %v\n, want %v", got.Snapshot, tt.want.Snapshot)
			}
		})
	}
}

func Test_collectAddOnIDs(t *testing.T) {
	var set = make(collection.StringSet)

	type args struct {
		dest   collection.StringSet
		addOns []AddOn
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Collection add-on ids",
			args: args{
				dest: set,
				addOns: []AddOn{
					stdAddOn,
					stdAddOn,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collectAddOnIDs(tt.args.dest, tt.args.addOns)

			assert.Equal(t, 1, len(tt.args.dest))

			t.Logf("%v", tt.args.dest)
		})
	}
}
