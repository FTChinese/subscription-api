package reader

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg"
	"github.com/FTChinese/subscription-api/pkg/addon"
	"github.com/FTChinese/subscription-api/pkg/invoice"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
)

func TestMembership_IsZero(t *testing.T) {
	type fields struct {
		MemberID      pkg.MemberID
		Edition       price.Edition
		LegacyTier    null.Int
		LegacyExpire  null.Int
		ExpireDate    chrono.Date
		PaymentMethod enum.PayMethod
		FtcPlanID     null.String
		StripeSubsID  null.String
		StripePlanID  null.String
		AutoRenewal   bool
		Status        enum.SubsStatus
		AppleSubsID   null.String
		B2BLicenceID  null.String
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name:   "Zero membership",
			fields: fields{},
			want:   true,
		},
		{
			name: "Non-zero membership",
			fields: fields{
				MemberID: pkg.MemberID{
					CompoundID: "",
					FtcID:      null.StringFrom(uuid.New().String()),
					UnionID:    null.String{},
				}.MustNormalize(),
				Edition: price.Edition{
					Tier:  enum.TierStandard,
					Cycle: enum.CycleYear,
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Membership{
				MemberID:      tt.fields.MemberID,
				Edition:       tt.fields.Edition,
				LegacyTier:    tt.fields.LegacyTier,
				LegacyExpire:  tt.fields.LegacyExpire,
				ExpireDate:    tt.fields.ExpireDate,
				PaymentMethod: tt.fields.PaymentMethod,
				FtcPlanID:     tt.fields.FtcPlanID,
				StripeSubsID:  tt.fields.StripeSubsID,
				StripePlanID:  tt.fields.StripePlanID,
				AutoRenewal:   tt.fields.AutoRenewal,
				Status:        tt.fields.Status,
				AppleSubsID:   tt.fields.AppleSubsID,
				B2BLicenceID:  tt.fields.B2BLicenceID,
			}
			if got := m.IsZero(); got != tt.want {
				t.Errorf("IsZero() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMembership_IsExpired(t *testing.T) {
	type fields struct {
		MemberID      pkg.MemberID
		Edition       price.Edition
		LegacyTier    null.Int
		LegacyExpire  null.Int
		ExpireDate    chrono.Date
		PaymentMethod enum.PayMethod
		FtcPlanID     null.String
		StripeSubsID  null.String
		StripePlanID  null.String
		AutoRenewal   bool
		Status        enum.SubsStatus
		AppleSubsID   null.String
		B2BLicenceID  null.String
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name:   "Empty member is expired",
			fields: fields{},
			want:   true,
		},
		{
			name: "Expired membership",
			fields: fields{
				MemberID: pkg.MemberID{
					CompoundID: "",
					FtcID:      null.StringFrom(uuid.New().String()),
					UnionID:    null.String{},
				}.MustNormalize(),
				ExpireDate: chrono.DateFrom(time.Now().AddDate(-1, 0, 0)),
			},
			want: true,
		},
		{
			name: "Stripe expired but auto renew",
			fields: fields{
				MemberID: pkg.MemberID{
					CompoundID: "",
					FtcID:      null.StringFrom(uuid.New().String()),
					UnionID:    null.String{},
				}.MustNormalize(),
				ExpireDate:   chrono.DateFrom(time.Now().AddDate(-1, 0, 0)),
				StripeSubsID: null.StringFrom(faker.GenStripeSubID()),
				StripePlanID: null.StringFrom(faker.GenStripePlanID()),
				AutoRenewal:  true,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Membership{
				MemberID:      tt.fields.MemberID,
				Edition:       tt.fields.Edition,
				LegacyTier:    tt.fields.LegacyTier,
				LegacyExpire:  tt.fields.LegacyExpire,
				ExpireDate:    tt.fields.ExpireDate,
				PaymentMethod: tt.fields.PaymentMethod,
				FtcPlanID:     tt.fields.FtcPlanID,
				StripeSubsID:  tt.fields.StripeSubsID,
				StripePlanID:  tt.fields.StripePlanID,
				AutoRenewal:   tt.fields.AutoRenewal,
				Status:        tt.fields.Status,
				AppleSubsID:   tt.fields.AppleSubsID,
				B2BLicenceID:  tt.fields.B2BLicenceID,
			}
			if got := m.IsExpired(); got != tt.want {
				t.Errorf("IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMembership_Normalize(t *testing.T) {
	type fields struct {
		MemberID      pkg.MemberID
		Edition       price.Edition
		LegacyTier    null.Int
		LegacyExpire  null.Int
		ExpireDate    chrono.Date
		PaymentMethod enum.PayMethod
		FtcPlanID     null.String
		StripeSubsID  null.String
		StripePlanID  null.String
		AutoRenewal   bool
		Status        enum.SubsStatus
		AppleSubsID   null.String
		B2BLicenceID  null.String
	}
	tests := []struct {
		name   string
		fields fields
		want   Membership
	}{
		{
			name: "Sync from legacy",
			fields: fields{
				MemberID: pkg.MemberID{
					CompoundID: "",
					FtcID:      null.StringFrom(uuid.New().String()),
					UnionID:    null.String{},
				}.MustNormalize(),
				Edition:       price.Edition{},
				LegacyTier:    null.IntFrom(10),
				LegacyExpire:  null.IntFrom(time.Now().Unix()),
				ExpireDate:    chrono.Date{},
				PaymentMethod: 0,
				FtcPlanID:     null.String{},
				StripeSubsID:  null.String{},
				StripePlanID:  null.String{},
				AutoRenewal:   false,
				Status:        0,
				AppleSubsID:   null.String{},
				B2BLicenceID:  null.String{},
			},
		},
		{
			name: "Sync to legacy",
			fields: fields{
				MemberID: pkg.MemberID{
					CompoundID: "",
					FtcID:      null.StringFrom(uuid.New().String()),
					UnionID:    null.String{},
				}.MustNormalize(),
				Edition: price.Edition{
					Tier:  enum.TierStandard,
					Cycle: enum.CycleYear,
				},
				LegacyTier:    null.Int{},
				LegacyExpire:  null.Int{},
				ExpireDate:    chrono.DateNow(),
				PaymentMethod: 0,
				FtcPlanID:     null.String{},
				StripeSubsID:  null.String{},
				StripePlanID:  null.String{},
				AutoRenewal:   false,
				Status:        0,
				AppleSubsID:   null.String{},
				B2BLicenceID:  null.String{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Membership{
				MemberID:      tt.fields.MemberID,
				Edition:       tt.fields.Edition,
				LegacyTier:    tt.fields.LegacyTier,
				LegacyExpire:  tt.fields.LegacyExpire,
				ExpireDate:    tt.fields.ExpireDate,
				PaymentMethod: tt.fields.PaymentMethod,
				FtcPlanID:     tt.fields.FtcPlanID,
				StripeSubsID:  tt.fields.StripeSubsID,
				StripePlanID:  tt.fields.StripePlanID,
				AutoRenewal:   tt.fields.AutoRenewal,
				Status:        tt.fields.Status,
				AppleSubsID:   tt.fields.AppleSubsID,
				B2BLicenceID:  tt.fields.B2BLicenceID,
			}
			got := m.Sync()

			assert.Equal(t, got.Tier, codeToTier[got.LegacyTier.Int64])
			assert.Equal(t, chrono.DateFrom(time.Unix(got.LegacyExpire.Int64, 0)).Time, got.ExpireDate.Time)
		})
	}
}

func TestMembership_WithinMaxRenewalPeriod(t *testing.T) {

	tests := []struct {
		name   string
		fields Membership
		want   bool
	}{
		{
			name: "Expire 1 one year can renew",
			fields: Membership{
				MemberID: pkg.MemberID{
					CompoundID: "",
					FtcID:      null.StringFrom(uuid.New().String()),
					UnionID:    null.String{},
				}.MustNormalize(),
				Edition: price.Edition{
					Tier:  enum.TierStandard,
					Cycle: enum.CycleYear,
				},
				ExpireDate:    chrono.DateFrom(time.Now().AddDate(1, 0, 0)),
				PaymentMethod: enum.PayMethodAli,
			},
			want: true,
		},
		{
			name: "Expire today can renew",
			fields: Membership{
				MemberID: pkg.MemberID{
					CompoundID: "",
					FtcID:      null.StringFrom(uuid.New().String()),
					UnionID:    null.String{},
				}.MustNormalize(),
				Edition: price.Edition{
					Tier:  enum.TierStandard,
					Cycle: enum.CycleYear,
				},
				ExpireDate:    chrono.DateFrom(time.Now()),
				PaymentMethod: enum.PayMethodAli,
			},
			want: true,
		},
		{
			name: "Expire on final date of 3rd year can renew",
			fields: Membership{
				MemberID: pkg.MemberID{
					CompoundID: "",
					FtcID:      null.StringFrom(uuid.New().String()),
					UnionID:    null.String{},
				}.MustNormalize(),
				Edition: price.Edition{
					Tier:  enum.TierStandard,
					Cycle: enum.CycleYear,
				},
				ExpireDate:    chrono.DateFrom(time.Now().AddDate(3, 0, 0)),
				PaymentMethod: enum.PayMethodAli,
			},
			want: true,
		},
		{
			name: "Expire 3+ years later cannot renew",
			fields: Membership{
				MemberID: pkg.MemberID{
					CompoundID: "",
					FtcID:      null.StringFrom(uuid.New().String()),
					UnionID:    null.String{},
				}.MustNormalize(),
				Edition: price.Edition{
					Tier:  enum.TierStandard,
					Cycle: enum.CycleYear,
				},
				ExpireDate:    chrono.DateFrom(time.Now().AddDate(3, 0, 1)),
				PaymentMethod: enum.PayMethodAli,
			},
			want: false,
		},
		{
			name: "Expired yesterday cannot renew",
			fields: Membership{
				MemberID: pkg.MemberID{
					CompoundID: "",
					FtcID:      null.StringFrom(uuid.New().String()),
					UnionID:    null.String{},
				}.MustNormalize(),
				Edition: price.Edition{
					Tier:  enum.TierStandard,
					Cycle: enum.CycleYear,
				},
				ExpireDate:    chrono.DateFrom(time.Now().AddDate(0, 0, -1)),
				PaymentMethod: enum.PayMethodAli,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.fields
			if got := m.WithinMaxRenewalPeriod(); got != tt.want {
				t.Errorf("canRenewViaAliWx() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMembership_RemainingDays(t *testing.T) {
	oneDayLater := time.Now().AddDate(0, 0, 1)

	hours := time.Until(oneDayLater).Hours()
	t.Logf("Hours: %f", hours)

	tomorrow := chrono.DateFrom(oneDayLater)
	remainHours := time.Until(tomorrow.Time).Hours()
	t.Logf("Remain hours: %f", remainHours)

	t.Logf("To days: %f", remainHours/24)

	m1 := Membership{
		ExpireDate: chrono.DateFrom(oneDayLater),
	}

	t.Logf("%d", m1.RemainingDays())

	m2 := Membership{
		ExpireDate: chrono.DateFrom(time.Now().AddDate(0, 0, -1)),
	}

	t.Logf("%d", m2.RemainingDays())
}

func TestMembership_WithInvoice(t *testing.T) {
	userID := uuid.New().String()

	current := NewMockMemberBuilder(userID).Build()

	type args struct {
		userID pkg.MemberID
		inv    invoice.Invoice
	}
	tests := []struct {
		name    string
		fields  Membership
		args    args
		want    Membership
		wantErr bool
	}{
		{
			name:   "Create membership",
			fields: Membership{},
			args: args{
				userID: pkg.NewFtcUserID(userID),
				inv:    invoice.NewMockInvoiceBuilder(userID).SetPeriodStart(time.Now()).Build(),
			},
			want: Membership{
				MemberID:      pkg.NewFtcUserID(userID),
				Edition:       faker.PriceStdYear.Edition,
				LegacyTier:    null.Int{},
				LegacyExpire:  null.Int{},
				ExpireDate:    chrono.DateFrom(time.Now().AddDate(1, 0, 1)),
				PaymentMethod: enum.PayMethodAli,
				FtcPlanID:     null.StringFrom(faker.PriceStdYear.ID),
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
		{
			name:   "Renew membership",
			fields: current,
			args: args{
				userID: pkg.NewFtcUserID(userID),
				inv: invoice.NewMockInvoiceBuilder(userID).
					WithOrderKind(enum.OrderKindRenew).
					SetPeriodStart(current.ExpireDate.Time).
					Build(),
			},
			want: Membership{
				MemberID:      pkg.NewFtcUserID(userID),
				Edition:       faker.PriceStdYear.Edition,
				LegacyTier:    null.Int{},
				LegacyExpire:  null.Int{},
				ExpireDate:    chrono.DateFrom(current.ExpireDate.AddDate(1, 0, 1)),
				PaymentMethod: enum.PayMethodAli,
				FtcPlanID:     null.StringFrom(faker.PriceStdYear.ID),
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
		{
			name:   "Upgrade membership",
			fields: current,
			args: args{
				userID: pkg.NewFtcUserID(userID),
				inv: invoice.NewMockInvoiceBuilder(userID).
					WithPrice(faker.PricePrm).
					WithOrderKind(enum.OrderKindUpgrade).
					SetPeriodStart(time.Now()).
					Build(),
			},
			want: Membership{
				MemberID:      pkg.NewFtcUserID(userID),
				Edition:       faker.PricePrm.Edition,
				LegacyTier:    null.Int{},
				LegacyExpire:  null.Int{},
				ExpireDate:    chrono.DateFrom(time.Now().AddDate(1, 0, 1)),
				PaymentMethod: enum.PayMethodAli,
				FtcPlanID:     null.StringFrom(faker.PricePrm.ID),
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
		{
			name:   "Membership addon",
			fields: current,
			args: args{
				userID: pkg.NewFtcUserID(userID),
				inv: invoice.NewMockInvoiceBuilder(userID).
					WithOrderKind(enum.OrderKindAddOn).
					Build(),
			},
			want: current.PlusAddOn(addon.AddOn{
				Standard: 367,
				Premium:  0,
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := tt.fields.WithInvoice(tt.args.userID, tt.args.inv)
			if (err != nil) != tt.wantErr {
				t.Errorf("WithInvoice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithInvoice() got = %v, want %v", got, tt.want)
			}
		})
	}
}
