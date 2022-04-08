package reader

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/addon"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/invoice"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/pw"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"reflect"
	"testing"
	"time"
)

func TestMembership_IsZero(t *testing.T) {
	type fields struct {
		MemberID      ids.UserIDs
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
				MemberID: ids.UserIDs{
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
				UserIDs:       tt.fields.MemberID,
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
		MemberID      ids.UserIDs
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
				MemberID: ids.UserIDs{
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
				MemberID: ids.UserIDs{
					CompoundID: "",
					FtcID:      null.StringFrom(uuid.New().String()),
					UnionID:    null.String{},
				}.MustNormalize(),
				ExpireDate:   chrono.DateFrom(time.Now().AddDate(-1, 0, 0)),
				StripeSubsID: null.StringFrom(faker.GenStripeSubID()),
				StripePlanID: null.StringFrom(faker.GenStripePriceID()),
				AutoRenewal:  true,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Membership{
				UserIDs:       tt.fields.MemberID,
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
		MemberID      ids.UserIDs
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
				MemberID: ids.UserIDs{
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
				MemberID: ids.UserIDs{
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
				UserIDs:       tt.fields.MemberID,
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

			t.Logf("%v", got)
			//assert.Equal(t, got.Tier, codeToTier[got.LegacyTier.Int64])
			//assert.Equal(t, chrono.DateFrom(time.Unix(got.LegacyExpire.Int64, 0)).Time, got.ExpireDate.Time)
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
				UserIDs: ids.UserIDs{
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
				UserIDs: ids.UserIDs{
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
				UserIDs: ids.UserIDs{
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
				UserIDs: ids.UserIDs{
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
				UserIDs: ids.UserIDs{
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
		userID ids.UserIDs
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
				userID: ids.NewFtcUserID(userID),
				inv:    invoice.Invoice{},
			},
			want: Membership{
				UserIDs:       ids.NewFtcUserID(userID),
				Edition:       pw.MockPwPriceStdYear.Edition,
				LegacyTier:    null.Int{},
				LegacyExpire:  null.Int{},
				ExpireDate:    chrono.DateFrom(time.Now().AddDate(1, 0, 1)),
				PaymentMethod: enum.PayMethodAli,
				FtcPlanID:     null.StringFrom(pw.MockPwPriceStdYear.ID),
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
				userID: ids.NewFtcUserID(userID),
				inv:    invoice.Invoice{},
			},
			want: Membership{
				UserIDs:       ids.NewFtcUserID(userID),
				Edition:       pw.MockPwPriceStdYear.Edition,
				LegacyTier:    null.Int{},
				LegacyExpire:  null.Int{},
				ExpireDate:    chrono.DateFrom(current.ExpireDate.AddDate(1, 0, 1)),
				PaymentMethod: enum.PayMethodAli,
				FtcPlanID:     null.StringFrom(pw.MockPwPriceStdYear.ID),
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
				userID: ids.NewFtcUserID(userID),
				inv:    invoice.Invoice{},
			},
			want: Membership{
				UserIDs:       ids.NewFtcUserID(userID),
				Edition:       pw.MockPwPricePrm.Edition,
				LegacyTier:    null.Int{},
				LegacyExpire:  null.Int{},
				ExpireDate:    chrono.DateFrom(time.Now().AddDate(1, 0, 1)),
				PaymentMethod: enum.PayMethodAli,
				FtcPlanID:     null.StringFrom(pw.MockPwPricePrm.ID),
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
				userID: ids.NewFtcUserID(userID),
				inv:    invoice.Invoice{},
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

func TestMembership_OrderKindByOneTime(t *testing.T) {

	userID := uuid.New().String()

	type args struct {
		e price.Edition
	}
	tests := []struct {
		name    string
		fields  Membership
		args    args
		want    enum.OrderKind
		wantErr bool
	}{
		{
			name:   "New member",
			fields: Membership{},
			args: args{
				e: price.StdYearEdition,
			},
			want:    enum.OrderKindCreate,
			wantErr: false,
		},
		{
			name:   "Renewal",
			fields: NewMockMemberBuilder(userID).Build(),
			args: args{
				e: price.StdYearEdition,
			},
			want:    enum.OrderKindRenew,
			wantErr: false,
		},
		{
			name:   "Exceed renewal",
			fields: NewMockMemberBuilder(userID).WithExpiration(time.Now().AddDate(3, 0, 1)).Build(),
			args: args{
				e: price.StdYearEdition,
			},
			want:    enum.OrderKindNull,
			wantErr: true,
		},
		{
			name:   "Upgrade",
			fields: NewMockMemberBuilder(userID).Build(),
			args: args{
				e: price.PremiumEdition,
			},
			want:    enum.OrderKindUpgrade,
			wantErr: false,
		},
		{
			name: "Standard add-on for premium",
			fields: NewMockMemberBuilder(userID).
				WithPrice(pw.MockPwPricePrm.FtcPrice).
				Build(),
			args: args{
				e: price.StdYearEdition,
			},
			want:    enum.OrderKindAddOn,
			wantErr: false,
		},
		{
			name:   "Stripe standard upgrade to premium",
			fields: NewMockMemberBuilder(userID).WithPayMethod(enum.PayMethodStripe).Build(),
			args: args{
				e: price.PremiumEdition,
			},
			want:    enum.OrderKindNull,
			wantErr: true,
		},
		{
			name: "Stripe standard purchase standard add-on",
			fields: NewMockMemberBuilder(userID).
				WithPayMethod(enum.PayMethodStripe).
				Build(),
			args: args{
				e: price.StdYearEdition,
			},
			want:    enum.OrderKindAddOn,
			wantErr: false,
		},
		{
			name: "Stripe premium purchase standard add-on",
			fields: NewMockMemberBuilder(userID).
				WithPayMethod(enum.PayMethodStripe).
				WithPrice(pw.MockPwPricePrm.FtcPrice).
				Build(),
			args: args{
				e: price.StdYearEdition,
			},
			want:    enum.OrderKindAddOn,
			wantErr: false,
		},
		{
			name: "Stripe premium purchase premium add-on",
			fields: NewMockMemberBuilder(userID).
				WithPayMethod(enum.PayMethodStripe).
				WithPrice(pw.MockPwPricePrm.FtcPrice).
				Build(),
			args: args{
				e: price.PremiumEdition,
			},
			want:    enum.OrderKindAddOn,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := tt.fields.OrderKindOfOneTime(tt.args.e)
			if (err != nil) != tt.wantErr {
				t.Errorf("OrderKindOfOneTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("OrderKindOfOneTime() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMembership_SubsKindByStripe(t *testing.T) {
	userID := uuid.New().String()

	type args struct {
		e price.Edition
	}
	tests := []struct {
		name    string
		fields  Membership
		args    args
		want    SubsKind
		wantErr bool
	}{
		{
			name:   "Empty",
			fields: Membership{},
			args: args{
				e: price.StdYearEdition,
			},
			want:    SubsKindCreate,
			wantErr: false,
		},
		{
			name:   "One-time standard switch to Stripe",
			fields: NewMockMemberBuilder(userID).Build(),
			args: args{
				e: price.StdYearEdition,
			},
			want:    SubsKindOneTimeToAutoRenew,
			wantErr: false,
		},
		{
			name: "Already Stripe Premium",
			fields: NewMockMemberBuilder(userID).
				WithPayMethod(enum.PayMethodStripe).
				WithPrice(pw.MockPwPricePrm.FtcPrice).
				Build(),
			args: args{
				e: price.StdYearEdition,
			},
			want:    SubsKindForbidden,
			wantErr: true,
		},
		{
			name:   "Stripe standard upgrade",
			fields: NewMockMemberBuilder(userID).WithPayMethod(enum.PayMethodStripe).Build(),
			args: args{
				e: price.PremiumEdition,
			},
			want:    SubsKindUpgrade,
			wantErr: false,
		},
		{
			name: "Stripe standard same cycle",
			fields: NewMockMemberBuilder(userID).
				WithPayMethod(enum.PayMethodStripe).
				Build(),
			args: args{
				e: price.StdYearEdition,
			},
			want:    SubsKindForbidden,
			wantErr: true,
		},
		{
			name: "Stripe standard different cycle",
			fields: NewMockMemberBuilder(userID).
				WithPayMethod(enum.PayMethodStripe).
				Build(),
			args: args{
				e: price.StdMonthEdition,
			},
			want:    SubsKindSwitchInterval,
			wantErr: false,
		},
		{
			name: "Active Apple cannot user Stripe",
			fields: NewMockMemberBuilder(userID).
				WithPayMethod(enum.PayMethodApple).
				Build(),
			args: args{
				e: price.StdMonthEdition,
			},
			want:    SubsKindForbidden,
			wantErr: true,
		},
		{
			name: "Active B2B cannot user Stripe",
			fields: NewMockMemberBuilder(userID).
				WithPayMethod(enum.PayMethodB2B).
				Build(),
			args: args{
				e: price.StdMonthEdition,
			},
			want:    SubsKindForbidden,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := tt.fields.SubsKindOfStripe(tt.args.e)
			if (err != nil) != tt.wantErr {
				t.Errorf("SubsKindOfStripe() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SubsKindOfStripe() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMembership_SubsKindByApple(t *testing.T) {
	userID := uuid.New().String()

	tests := []struct {
		name    string
		fields  Membership
		want    SubsKind
		wantErr bool
	}{
		{
			name:    "Empty",
			fields:  Membership{},
			want:    SubsKindCreate,
			wantErr: false,
		},
		{
			name:    "One-time carried over",
			fields:  NewMockMemberBuilder(userID).Build(),
			want:    SubsKindOneTimeToAutoRenew,
			wantErr: false,
		},
		{
			name:    "Active Stripe cannot use Apple",
			fields:  NewMockMemberBuilder(userID).WithPayMethod(enum.PayMethodStripe).Build(),
			want:    SubsKindForbidden,
			wantErr: true,
		},
		{
			name:    "Active Apple can be refreshed",
			fields:  NewMockMemberBuilder(userID).WithPayMethod(enum.PayMethodApple).Build(),
			want:    SubsKindRefreshAutoRenew,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := tt.fields.SubsKindByApple()
			if (err != nil) != tt.wantErr {
				t.Errorf("SubsKindByApple() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SubsKindByApple() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMembership_OfferKindsEnjoyed(t *testing.T) {

	tests := []struct {
		name   string
		fields Membership
		want   []price.OfferKind
	}{
		{
			name:   "New member enjoys only promotion offer",
			fields: Membership{},
			want:   []price.OfferKind{price.OfferKindPromotion},
		},
		{
			name:   "Existing member enjoys promotion and retention",
			fields: NewMockMemberBuilder("").Build(),
			want: []price.OfferKind{
				price.OfferKindPromotion,
				price.OfferKindRetention,
			},
		},
		{
			name: "Expired member enjoys promotion and win-back",
			fields: NewMockMemberBuilder("").
				WithExpiration(time.Now().AddDate(0, -1, 0)).
				Build(),
			want: []price.OfferKind{
				price.OfferKindPromotion,
				price.OfferKindWinBack,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if got := tt.fields.OfferKindsEnjoyed(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("OfferKindsEnjoyed() = %v, want %v", got, tt.want)
			}
		})
	}
}
