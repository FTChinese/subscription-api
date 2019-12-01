package subscription

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
	"gitlab.com/ftchinese/subscription-api/models/plan"
	"gitlab.com/ftchinese/subscription-api/models/reader"
	"testing"
	"time"
)

func TestGenerateSnapshotID(t *testing.T) {
	t.Log(GenerateSnapshotID())
}

func TestMembership_IsExpired(t *testing.T) {
	profile := NewProfile()

	type fields struct {
		ID            null.String
		AccountID     reader.MemberID
		LegacyTier    null.Int
		LegacyExpire  null.Int
		Coordinate    plan.BasePlan
		ExpireDate    chrono.Date
		PaymentMethod enum.PayMethod
		StripeSubID   null.String
		StripePlanID  null.String
		AutoRenewal   bool
		Status        SubStatus
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name:   "Empty Member",
			fields: fields{},
			want:   true,
		},
		{
			name: "Expired",
			fields: fields{
				AccountID: profile.AccountID(reader.AccountKindFtc),
				Coordinate: plan.BasePlan{
					Tier:  enum.TierStandard,
					Cycle: enum.CycleYear,
				},
				ExpireDate: chrono.DateFrom(time.Now().AddDate(0, -1, 0)),
			},
			want: true,
		},
		{
			name: "Expired but auto renewal",
			fields: fields{
				AccountID: profile.AccountID(reader.AccountKindFtc),
				Coordinate: plan.BasePlan{
					Tier:  enum.TierStandard,
					Cycle: enum.CycleYear,
				},
				ExpireDate:  chrono.DateFrom(time.Now().AddDate(0, -1, 0)),
				AutoRenewal: true,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Membership{
				ID:            tt.fields.ID,
				MemberID:      tt.fields.AccountID,
				LegacyTier:    tt.fields.LegacyTier,
				LegacyExpire:  tt.fields.LegacyExpire,
				BasePlan:      tt.fields.Coordinate,
				ExpireDate:    tt.fields.ExpireDate,
				PaymentMethod: tt.fields.PaymentMethod,
				StripeSubID:   tt.fields.StripeSubID,
				StripePlanID:  tt.fields.StripePlanID,
				AutoRenewal:   tt.fields.AutoRenewal,
				Status:        tt.fields.Status,
			}
			if got := m.IsExpired(); got != tt.want {
				t.Errorf("IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMembership_inRenewalPeriod(t *testing.T) {
	type fields struct {
		ID            null.String
		AccountID     reader.MemberID
		LegacyTier    null.Int
		LegacyExpire  null.Int
		Coordinate    plan.BasePlan
		ExpireDate    chrono.Date
		PaymentMethod enum.PayMethod
		StripeSubID   null.String
		StripePlanID  null.String
		AutoRenewal   bool
		Status        SubStatus
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name:   "Expired Member",
			fields: fields{},
			want:   false,
		},
		{
			name: "Exceed",
			fields: fields{
				ExpireDate: chrono.DateFrom(time.Now().AddDate(3, 1, 0)),
			},
			want: false,
		},
		{
			name: "Expired",
			fields: fields{
				ExpireDate: chrono.DateFrom(time.Now().AddDate(0, -1, 0)),
			},
			want: false,
		},
		{
			name: "Expires on Today",
			fields: fields{
				ExpireDate: chrono.DateFrom(time.Now()),
			},
			want: true,
		},
		{
			name: "Expires three years later",
			fields: fields{
				ExpireDate: chrono.DateFrom(time.Now().AddDate(3, 0, 0)),
			},
			want: true,
		},
		{
			name: "Renewal",
			fields: fields{
				ExpireDate: chrono.DateFrom(time.Now().AddDate(1, 0, 0)),
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Membership{
				ID:            tt.fields.ID,
				MemberID:      tt.fields.AccountID,
				LegacyTier:    tt.fields.LegacyTier,
				LegacyExpire:  tt.fields.LegacyExpire,
				BasePlan:      tt.fields.Coordinate,
				ExpireDate:    tt.fields.ExpireDate,
				PaymentMethod: tt.fields.PaymentMethod,
				StripeSubID:   tt.fields.StripeSubID,
				StripePlanID:  tt.fields.StripePlanID,
				AutoRenewal:   tt.fields.AutoRenewal,
				Status:        tt.fields.Status,
			}
			if got := m.inRenewalPeriod(); got != tt.want {
				t.Errorf("inRenewalPeriod() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMembership_PermitRenewal(t *testing.T) {
	type fields struct {
		ID            null.String
		AccountID     reader.MemberID
		LegacyTier    null.Int
		LegacyExpire  null.Int
		Coordinate    plan.BasePlan
		ExpireDate    chrono.Date
		PaymentMethod enum.PayMethod
		StripeSubID   null.String
		StripePlanID  null.String
		AutoRenewal   bool
		Status        SubStatus
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name:   "Empty member",
			fields: fields{},
			want:   false,
		},
		{
			name: "Auto renewal",
			fields: fields{
				AutoRenewal: true,
			},
			want: false,
		},
		{
			name: "Not Ali or Wx",
			fields: fields{
				PaymentMethod: enum.PayMethodStripe,
			},
			want: false,
		},
		{
			name: "Beyond renewal",
			fields: fields{
				ExpireDate: chrono.DateFrom(time.Now().AddDate(3, 1, 0)),
			},
			want: false,
		},
		{
			name: "Allow Renewal",
			fields: fields{
				PaymentMethod: enum.PayMethodAli,
				ExpireDate:    chrono.DateFrom(time.Now().AddDate(1, 0, 0)),
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Membership{
				ID:            tt.fields.ID,
				MemberID:      tt.fields.AccountID,
				LegacyTier:    tt.fields.LegacyTier,
				LegacyExpire:  tt.fields.LegacyExpire,
				BasePlan:      tt.fields.Coordinate,
				ExpireDate:    tt.fields.ExpireDate,
				PaymentMethod: tt.fields.PaymentMethod,
				StripeSubID:   tt.fields.StripeSubID,
				StripePlanID:  tt.fields.StripePlanID,
				AutoRenewal:   tt.fields.AutoRenewal,
				Status:        tt.fields.Status,
			}
			if got := m.PermitRenewal(); got != tt.want {
				t.Errorf("PermitRenewal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMembership_SubsKind(t *testing.T) {

	profile := NewProfile()

	type fields struct {
		ID            null.String
		AccountID     reader.MemberID
		LegacyTier    null.Int
		LegacyExpire  null.Int
		Coordinate    plan.BasePlan
		ExpireDate    chrono.Date
		PaymentMethod enum.PayMethod
		StripeSubID   null.String
		StripePlanID  null.String
		AutoRenewal   bool
		Status        SubStatus
	}
	type args struct {
		p plan.Plan
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    SubsKind
		wantErr bool
	}{
		{
			name:   "Empty member",
			fields: fields{},
			args: args{
				p: plan.standardYearlyPlan,
			},
			want:    SubsKindCreate,
			wantErr: false,
		},
		{
			name: "Invalid Stripe Subscription",
			fields: fields{
				Status: SubStatusIncompleteExpired,
			},
			args:    args{p: plan.standardYearlyPlan},
			want:    SubsKindCreate,
			wantErr: false,
		},
		{
			name: "Expired subscription",
			fields: fields{
				ExpireDate: chrono.DateFrom(time.Now().AddDate(1, -1, 0)),
			},
			args:    args{p: plan.standardYearlyPlan},
			want:    SubsKindCreate,
			wantErr: false,
		},
		{
			name: "Renewal",
			fields: fields{
				AccountID: profile.AccountID(reader.AccountKindFtc),
				Coordinate: plan.BasePlan{
					Tier:  enum.TierStandard,
					Cycle: enum.CycleYear,
				},
				ExpireDate: chrono.DateFrom(time.Now().AddDate(1, 0, 0)),
			},
			args:    args{p: plan.standardYearlyPlan},
			want:    SubsKindRenew,
			wantErr: false,
		},
		{
			name: "Upgrading",
			fields: fields{
				AccountID: profile.AccountID(reader.AccountKindFtc),
				Coordinate: plan.BasePlan{
					Tier:  enum.TierStandard,
					Cycle: enum.CycleYear,
				},
				ExpireDate: chrono.DateFrom(time.Now().AddDate(1, 0, 0)),
			},
			args:    args{p: plan.premiumYearlyPlan},
			want:    SubsKindUpgrade,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Membership{
				ID:            tt.fields.ID,
				MemberID:      tt.fields.AccountID,
				LegacyTier:    tt.fields.LegacyTier,
				LegacyExpire:  tt.fields.LegacyExpire,
				BasePlan:      tt.fields.Coordinate,
				ExpireDate:    tt.fields.ExpireDate,
				PaymentMethod: tt.fields.PaymentMethod,
				StripeSubID:   tt.fields.StripeSubID,
				StripePlanID:  tt.fields.StripePlanID,
				AutoRenewal:   tt.fields.AutoRenewal,
				Status:        tt.fields.Status,
			}
			got, err := m.SubsKind(tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("SubsKind() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SubsKind() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMembership_PermitStripeCreate(t *testing.T) {
	profile := NewProfile()

	type fields struct {
		ID            null.String
		AccountID     reader.MemberID
		LegacyTier    null.Int
		LegacyExpire  null.Int
		Coordinate    plan.BasePlan
		ExpireDate    chrono.Date
		PaymentMethod enum.PayMethod
		StripeSubID   null.String
		StripePlanID  null.String
		AutoRenewal   bool
		Status        SubStatus
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:    "Empty Member",
			fields:  fields{},
			wantErr: false,
		},

		{
			name: "Expired Alipay",
			fields: fields{
				AccountID: profile.AccountID(reader.AccountKindFtc),
				Coordinate: plan.BasePlan{
					Tier:  enum.TierStandard,
					Cycle: enum.CycleYear,
				},
				ExpireDate: chrono.DateFrom(time.Now().AddDate(0, -1, 0)),
			},
			wantErr: false,
		},
		{
			name: "Not Expired Wx",
			fields: fields{
				AccountID: profile.AccountID(reader.AccountKindFtc),
				Coordinate: plan.BasePlan{
					Tier:  enum.TierStandard,
					Cycle: enum.CycleYear,
				},
				ExpireDate: chrono.DateFrom(time.Now().AddDate(1, 0, 0)),
			},
			wantErr: true,
		},
		{
			name: "Valid Stripe",
			fields: fields{
				AccountID:     profile.AccountID(reader.AccountKindFtc),
				ExpireDate:    chrono.DateFrom(time.Now().AddDate(1, 0, 0)),
				PaymentMethod: enum.PayMethodStripe,
				AutoRenewal:   true,
				Status:        SubStatusActive,
			},
			wantErr: true,
		},
		{
			name: "Stripe Incomplete Expired",
			fields: fields{
				AccountID:     profile.AccountID(reader.AccountKindFtc),
				ExpireDate:    chrono.DateFrom(time.Now().AddDate(1, 0, 0)),
				PaymentMethod: enum.PayMethodStripe,
				AutoRenewal:   true,
				Status:        SubStatusIncompleteExpired,
			},
			wantErr: false,
		},
		{
			name: "Expired stripe but auto renewal",
			fields: fields{
				AccountID:     profile.AccountID(reader.AccountKindFtc),
				ExpireDate:    chrono.DateFrom(time.Now().AddDate(0, -1, 0)),
				PaymentMethod: enum.PayMethodStripe,
				AutoRenewal:   true,
				Status:        SubStatusActive,
			},
			wantErr: true,
		},
		{
			name: "Expired stripe auto renewal but inactive",
			fields: fields{
				AccountID:     profile.AccountID(reader.AccountKindFtc),
				ExpireDate:    chrono.DateFrom(time.Now().AddDate(0, -1, 0)),
				PaymentMethod: enum.PayMethodStripe,
				AutoRenewal:   true,
				Status:        SubStatusIncompleteExpired,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Membership{
				ID:            tt.fields.ID,
				MemberID:      tt.fields.AccountID,
				LegacyTier:    tt.fields.LegacyTier,
				LegacyExpire:  tt.fields.LegacyExpire,
				BasePlan:      tt.fields.Coordinate,
				ExpireDate:    tt.fields.ExpireDate,
				PaymentMethod: tt.fields.PaymentMethod,
				StripeSubID:   tt.fields.StripeSubID,
				StripePlanID:  tt.fields.StripePlanID,
				AutoRenewal:   tt.fields.AutoRenewal,
				Status:        tt.fields.Status,
			}
			if err := m.PermitStripeCreate(); (err != nil) != tt.wantErr {
				t.Errorf("PermitStripeCreate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
