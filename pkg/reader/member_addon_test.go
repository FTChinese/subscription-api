package reader

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/pkg/addon"
	"github.com/FTChinese/subscription-api/pkg/invoice"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"reflect"
	"testing"
	"time"
)

func TestMembership_withAddOnInvoice(t *testing.T) {

	userID := uuid.New().String()

	inv := invoice.NewMockInvoiceBuilder().
		WithFtcID(userID).
		WithOrderKind(enum.OrderKindAddOn).
		Build().
		SetPeriod(time.Now())

	current := NewMockMemberBuilderV2(enum.AccountKindFtc).
		WithFtcID(userID).
		WithExpiration(time.Now().AddDate(0, 0, -1)).
		Build().
		PlusAddOn(addon.New(inv.Tier, inv.TotalDays()))

	type args struct {
		i invoice.Invoice
	}
	tests := []struct {
		name    string
		fields  Membership
		args    args
		want    Membership
		wantErr bool
	}{
		{
			name:   "Transfer add on",
			fields: current,
			args: args{
				i: inv,
			},
			want: Membership{
				UserIDs:       current.UserIDs,
				Edition:       inv.Edition,
				LegacyTier:    null.Int{},
				LegacyExpire:  null.Int{},
				ExpireDate:    chrono.DateFrom(inv.EndUTC.Time),
				PaymentMethod: inv.PaymentMethod,
				FtcPlanID:     inv.PriceID,
				StripeSubsID:  null.String{},
				StripePlanID:  null.String{},
				AutoRenewal:   false,
				Status:        0,
				AppleSubsID:   null.String{},
				B2BLicenceID:  null.String{},
				AddOn:         addon.AddOn{},
			}.Sync(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := tt.fields.withAddOnInvoice(tt.args.i)
			if (err != nil) != tt.wantErr {
				t.Errorf("withAddOnInvoice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("withAddOnInvoice() got = %v, \nwant %v", got, tt.want)
			}
		})
	}
}

func TestMembership_addonToInvoice(t *testing.T) {

	m := NewMockMemberBuilderV2(enum.AccountKindFtc).
		WithAddOn(addon.AddOn{
			Standard: 123,
			Premium:  0,
		}).
		Build()

	tests := []struct {
		name   string
		fields Membership
		want   invoice.Invoice
	}{
		{
			name:   "Turn addon to invoice",
			fields: m,
			want: invoice.Invoice{
				ID:         "",
				CompoundID: m.CompoundID,
				Edition: price.Edition{
					Tier:  enum.TierStandard,
					Cycle: enum.CycleYear,
				},
				YearMonthDay: dt.YearMonthDay{
					Years:  0,
					Months: 0,
					Days:   123,
				},
				AddOnSource:   addon.SourceCarryOver,
				AppleTxID:     null.String{},
				OrderID:       null.String{},
				OrderKind:     enum.OrderKindAddOn,
				PaidAmount:    0,
				PaymentMethod: enum.PayMethodAli,
				PriceID:       null.String{},
				StripeSubsID:  null.String{},
				CreatedUTC:    chrono.TimeNow(),
				ConsumedUTC:   chrono.TimeNow(),
				DateTimePeriod: dt.DateTimePeriod{
					StartUTC: chrono.TimeFrom(m.ExpireDate.Time),
					EndUTC:   chrono.TimeFrom(m.ExpireDate.AddDate(0, 0, 123)),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.fields
			got := m.addonToInvoice()

			tt.want.ID = got.ID
			tt.want.CreatedUTC = got.CreatedUTC
			tt.want.ConsumedUTC = got.ConsumedUTC

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("addonToInvoice() = %v, \nwant %v", got, tt.want)

				return
			}

			t.Logf("%+v", got)
		})
	}
}

func TestMembership_pickConsumableAddOn(t *testing.T) {

	ftcID := uuid.New().String()

	type args struct {
		groupedInv invoice.AddOnGroup
	}
	tests := []struct {
		name   string
		fields Membership
		args   args
		want   int
	}{
		{
			name: "Use membership addon",
			fields: NewMockMemberBuilderV2(enum.AccountKindFtc).
				WithAddOn(addon.AddOn{
					Standard: 1111,
					Premium:  0,
				}).Build(),
			args: args{
				groupedInv: map[enum.Tier][]invoice.Invoice{
					enum.TierStandard: nil,
					enum.TierPremium:  nil,
				},
			},
			want: 1,
		},
		{
			name: "User invoice addon",
			fields: NewMockMemberBuilderV2(enum.AccountKindFtc).
				Build(),
			args: args{
				groupedInv: map[enum.Tier][]invoice.Invoice{
					enum.TierStandard: {
						invoice.NewMockInvoiceBuilder().
							WithFtcID(ftcID).
							WithOrderKind(enum.OrderKindAddOn).
							WithAddOnSource(addon.SourceUserPurchase).
							Build(),
						invoice.NewMockInvoiceBuilder().
							WithFtcID(ftcID).
							WithOrderKind(enum.OrderKindAddOn).
							WithAddOnSource(addon.SourceCompensation).
							Build(),
					},
					enum.TierPremium: nil,
				},
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.fields
			got := m.pickConsumableAddOn(tt.args.groupedInv)

			t.Logf("%+v", got)

			if len(got) != tt.want {
				t.Errorf("pickConsumableAddOn() = %v, want %v", len(got), tt.want)
				return
			}
		})
	}
}

func TestMembership_ClaimAddOns(t *testing.T) {

	userID := uuid.New().String()

	inv := invoice.NewMockInvoiceBuilder().
		WithFtcID(userID).
		WithOrderKind(enum.OrderKindAddOn).
		Build().
		SetPeriod(time.Now())

	m := NewMockMemberBuilderV2(enum.AccountKindFtc).
		WithFtcID(userID).
		WithExpiration(time.Now().AddDate(0, 0, -1)).
		Build().
		PlusAddOn(addon.New(inv.Tier, inv.TotalDays()))

	type args struct {
		inv []invoice.Invoice
	}
	tests := []struct {
		name   string
		fields Membership
		args   args
		//want    AddOnClaimed
		wantErr bool
	}{
		{
			name:   "Transfer addon invoices",
			fields: m,
			args: args{
				inv: []invoice.Invoice{
					invoice.NewMockInvoiceBuilder().
						WithFtcID(userID).
						WithOrderKind(enum.OrderKindAddOn).
						Build(),
					invoice.NewMockInvoiceBuilder().
						WithFtcID(userID).
						WithPrice(price.MockPricePrm).
						WithOrderKind(enum.OrderKindAddOn).
						Build(),
				},
			},
			wantErr: false,
		},
		{
			name: "Transfer addon invoices",
			fields: NewMockMemberBuilderV2(enum.AccountKindFtc).
				WithFtcID(userID).
				WithExpiration(time.Now().AddDate(0, 0, -1)).
				WithAddOn(addon.AddOn{
					Standard: 1000,
					Premium:  0,
				}).
				Build(),
			args: args{
				inv: []invoice.Invoice{
					invoice.NewMockInvoiceBuilder().
						WithFtcID(userID).
						WithOrderKind(enum.OrderKindAddOn).
						Build(),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.fields

			got, err := m.ClaimAddOns(tt.args.inv)
			if (err != nil) != tt.wantErr {
				t.Errorf("ClaimAddOns() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("ClaimAddOns() got = %v, want %v", got, tt.want)
			//}

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}
