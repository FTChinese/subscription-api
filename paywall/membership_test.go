package paywall

import (
	"reflect"
	"testing"
	"time"

	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
)

func TestMembership_FromStripe(t *testing.T) {
	profile := NewProfile()

	s := GetStripeSub()

	type fields struct {
		member Membership
	}
	type args struct {
		id  AccountID
		sub StripeSub
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "New member",
			fields: fields{
				member: Membership{},
			},
			args: args{
				id:  profile.UserID(AccountKindFtc),
				sub: NewStripeSub(&s),
			},
			wantErr: false,
		},
		{
			name: "Existing member",
			fields: fields{
				member: profile.Membership(AccountKindLinked, enum.PayMethodAli, true),
			},
			args: args{
				id:  profile.UserID(AccountKindLinked),
				sub: NewStripeSub(&s),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.fields.member

			got, err := m.FromStripe(tt.args.id, tt.args.sub)
			if (err != nil) != tt.wantErr {
				t.Errorf("Membership.FromStripe() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("Stripe member: %+v", got)
		})
	}
}

func TestMembership_FromAliOrWx(t *testing.T) {
	profile := NewProfile()

	type fields struct {
		member Membership
	}
	type args struct {
		sub Subscription
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Create Member",
			fields: fields{
				member: Membership{},
			},
			args: args{
				sub: profile.AliWxSub(AccountKindFtc, enum.PayMethodWx, SubsKindCreate),
			},
			wantErr: false,
		},
		{
			name: "Renew Member",
			fields: fields{
				member: profile.Membership(AccountKindFtc, enum.PayMethodWx, false),
			},
			args: args{
				sub: profile.AliWxSub(AccountKindFtc, enum.PayMethodAli, SubsKindRenew),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.fields.member

			got, err := m.FromAliOrWx(tt.args.sub)
			if (err != nil) != tt.wantErr {
				t.Errorf("Membership.FromAliOrWx() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("%+v", got)
		})
	}
}

func TestMembership_IsRenewAllowed(t *testing.T) {
	profile := NewProfile()
	m := profile.Membership(AccountKindFtc, enum.PayMethodAli, false)

	type fields struct {
		ID            null.String
		UserID        AccountID
		Coordinate    Coordinate
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
			name: "Renew Allowed",
			fields: fields{
				ID: m.ID,
				UserID: AccountID{
					CompoundID: m.CompoundID,
					FtcID:      m.FtcID,
					UnionID:    m.UnionID,
				},
				Coordinate: Coordinate{
					Tier:  m.Tier,
					Cycle: m.Cycle,
				},
				ExpireDate:    m.ExpireDate,
				PaymentMethod: m.PaymentMethod,
				StripeSubID:   m.StripeSubID,
				StripePlanID:  m.StripePlanID,
				AutoRenewal:   m.AutoRenewal,
				Status:        m.Status,
			},
			want: true,
		},
		{
			name: "Beyond Renewal",
			fields: fields{
				ID: m.ID,
				UserID: AccountID{
					CompoundID: m.CompoundID,
					FtcID:      m.FtcID,
					UnionID:    m.UnionID,
				},
				Coordinate: Coordinate{
					Tier:  m.Tier,
					Cycle: m.Cycle,
				},
				ExpireDate:    chrono.DateFrom(time.Now().AddDate(3, 0, 1)),
				PaymentMethod: m.PaymentMethod,
				StripeSubID:   m.StripeSubID,
				StripePlanID:  m.StripePlanID,
				AutoRenewal:   m.AutoRenewal,
				Status:        m.Status,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Membership{
				ID:            tt.fields.ID,
				AccountID:     tt.fields.UserID,
				Coordinate:    tt.fields.Coordinate,
				ExpireDate:    tt.fields.ExpireDate,
				PaymentMethod: tt.fields.PaymentMethod,
				StripeSubID:   tt.fields.StripeSubID,
				StripePlanID:  tt.fields.StripePlanID,
				AutoRenewal:   tt.fields.AutoRenewal,
				Status:        tt.fields.Status,
			}
			if got := m.IsRenewAllowed(); got != tt.want {
				t.Errorf("Membership.IsRenewAllowed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMembership_IsExpired(t *testing.T) {
	profile := NewProfile()
	m := profile.Membership(AccountKindFtc, enum.PayMethodAli, false)

	type fields struct {
		ID            null.String
		UserID        AccountID
		Coordinate    Coordinate
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
			name: "Expired Member",
			fields: fields{
				ID: m.ID,
				UserID: AccountID{
					CompoundID: m.CompoundID,
					FtcID:      m.FtcID,
					UnionID:    m.UnionID,
				},
				Coordinate: Coordinate{
					Tier:  m.Tier,
					Cycle: m.Cycle,
				},
				ExpireDate:    chrono.DateFrom(time.Now().AddDate(0, -1, 0)),
				PaymentMethod: m.PaymentMethod,
				StripeSubID:   m.StripeSubID,
				StripePlanID:  m.StripePlanID,
				AutoRenewal:   m.AutoRenewal,
				Status:        m.Status,
			},
			want: true,
		},
		{
			name: "Not Expired Member",
			fields: fields{
				ID: m.ID,
				UserID: AccountID{
					CompoundID: m.CompoundID,
					FtcID:      m.FtcID,
					UnionID:    m.UnionID,
				},
				Coordinate: Coordinate{
					Tier:  m.Tier,
					Cycle: m.Cycle,
				},
				ExpireDate:    m.ExpireDate,
				PaymentMethod: m.PaymentMethod,
				StripeSubID:   m.StripeSubID,
				StripePlanID:  m.StripePlanID,
				AutoRenewal:   m.AutoRenewal,
				Status:        m.Status,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Membership{
				ID:            tt.fields.ID,
				AccountID:     tt.fields.UserID,
				Coordinate:    tt.fields.Coordinate,
				ExpireDate:    tt.fields.ExpireDate,
				PaymentMethod: tt.fields.PaymentMethod,
				StripeSubID:   tt.fields.StripeSubID,
				StripePlanID:  tt.fields.StripePlanID,
				AutoRenewal:   tt.fields.AutoRenewal,
				Status:        tt.fields.Status,
			}
			if got := m.IsExpired(); got != tt.want {
				t.Errorf("Membership.IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMembership_PermitStripeCreate(t *testing.T) {
	profile := NewProfile()
	m := profile.Membership(AccountKindFtc, enum.PayMethodStripe, false)
	m.Status = SubStatusIncomplete

	type fields struct {
		member Membership
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Empty member",
			fields: fields{
				member: Membership{},
			},
			wantErr: false,
		},
		{
			name: "Expired Alipay",
			fields: fields{
				member: profile.Membership(AccountKindFtc, enum.PayMethodAli, true),
			},
			wantErr: false,
		},
		{
			name: "Not expired wxpay",
			fields: fields{
				member: profile.Membership(AccountKindFtc, enum.PayMethodWx, false),
			},
			wantErr: true,
		},
		{
			name: "Active Stripe User",
			fields: fields{
				member: profile.Membership(AccountKindFtc, enum.PayMethodStripe, false),
			},
			wantErr: true,
		},
		{
			name: "Incomplete Stripe User",
			fields: fields{
				member: m,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.fields.member

			if err := m.PermitStripeCreate(); (err != nil) != tt.wantErr {
				t.Errorf("Membership.PermitStripeCreate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_genMmID(t *testing.T) {
	tests := []struct {
		name    string
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateMemberID()
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateMemberID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GenerateMemberID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewMember(t *testing.T) {
	type args struct {
		u AccountID
	}
	tests := []struct {
		name string
		args args
		want Membership
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewMember(tt.args.u); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMember() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMembership_FromGiftCard(t *testing.T) {
	type fields struct {
		ID            null.String
		UserID        AccountID
		Coordinate    Coordinate
		ExpireDate    chrono.Date
		PaymentMethod enum.PayMethod
		StripeSubID   null.String
		StripePlanID  null.String
		AutoRenewal   bool
		Status        SubStatus
	}
	type args struct {
		c GiftCard
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Membership
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Membership{
				ID:            tt.fields.ID,
				AccountID:     tt.fields.UserID,
				Coordinate:    tt.fields.Coordinate,
				ExpireDate:    tt.fields.ExpireDate,
				PaymentMethod: tt.fields.PaymentMethod,
				StripeSubID:   tt.fields.StripeSubID,
				StripePlanID:  tt.fields.StripePlanID,
				AutoRenewal:   tt.fields.AutoRenewal,
				Status:        tt.fields.Status,
			}
			got, err := m.FromGiftCard(tt.args.c)
			if (err != nil) != tt.wantErr {
				t.Errorf("Membership.FromGiftCard() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Membership.FromGiftCard() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMembership_IsZero(t *testing.T) {
	type fields struct {
		ID            null.String
		UserID        AccountID
		Coordinate    Coordinate
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Membership{
				ID:            tt.fields.ID,
				AccountID:     tt.fields.UserID,
				Coordinate:    tt.fields.Coordinate,
				ExpireDate:    tt.fields.ExpireDate,
				PaymentMethod: tt.fields.PaymentMethod,
				StripeSubID:   tt.fields.StripeSubID,
				StripePlanID:  tt.fields.StripePlanID,
				AutoRenewal:   tt.fields.AutoRenewal,
				Status:        tt.fields.Status,
			}
			if got := m.IsZero(); got != tt.want {
				t.Errorf("Membership.IsZero() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMembership_IsAliOrWxPay(t *testing.T) {
	type fields struct {
		ID            null.String
		UserID        AccountID
		Coordinate    Coordinate
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Membership{
				ID:            tt.fields.ID,
				AccountID:     tt.fields.UserID,
				Coordinate:    tt.fields.Coordinate,
				ExpireDate:    tt.fields.ExpireDate,
				PaymentMethod: tt.fields.PaymentMethod,
				StripeSubID:   tt.fields.StripeSubID,
				StripePlanID:  tt.fields.StripePlanID,
				AutoRenewal:   tt.fields.AutoRenewal,
				Status:        tt.fields.Status,
			}
			if got := m.IsAliOrWxPay(); got != tt.want {
				t.Errorf("Membership.IsAliOrWxPay() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMembership_SubsKind(t *testing.T) {
	type fields struct {
		ID            null.String
		UserID        AccountID
		Coordinate    Coordinate
		ExpireDate    chrono.Date
		PaymentMethod enum.PayMethod
		StripeSubID   null.String
		StripePlanID  null.String
		AutoRenewal   bool
		Status        SubStatus
	}
	type args struct {
		p Plan
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    SubsKind
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Membership{
				ID:            tt.fields.ID,
				AccountID:     tt.fields.UserID,
				Coordinate:    tt.fields.Coordinate,
				ExpireDate:    tt.fields.ExpireDate,
				PaymentMethod: tt.fields.PaymentMethod,
				StripeSubID:   tt.fields.StripeSubID,
				StripePlanID:  tt.fields.StripePlanID,
				AutoRenewal:   tt.fields.AutoRenewal,
				Status:        tt.fields.Status,
			}
			got, err := m.SubsKind(tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("Membership.SubsKind() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Membership.SubsKind() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMembership_PermitStripeUpgrade(t *testing.T) {

	profile := NewProfile()
	incompleteM := profile.Membership(AccountKindFtc, enum.PayMethodStripe, false)
	incompleteM.Status = SubStatusIncomplete

	p := StripeSubParams{
		Coordinate: Coordinate{
			Tier:  enum.TierPremium,
			Cycle: enum.CycleYear,
		},
	}

	type fields struct {
		member Membership
	}
	type args struct {
		p StripeSubParams
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "Empty member",
			fields: fields{
				member: Membership{},
			},
			args: args{
				p: p,
			},
			want: false,
		},
		{
			name: "Non stripe member",
			fields: fields{
				member: profile.Membership(AccountKindFtc, enum.PayMethodAli, false),
			},
			args: args{
				p: p,
			},
			want: false,
		},
		{
			name: "Not active stripe member",
			fields: fields{
				member: incompleteM,
			},
			args: args{
				p: p,
			},
			want: false,
		},
		{
			name: "Stripe active standard member",
			fields: fields{
				member: profile.Membership(AccountKindFtc, enum.PayMethodStripe, false),
			},
			args: args{
				p: p,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.fields.member

			if got := m.PermitStripeUpgrade(tt.args.p); got != tt.want {
				t.Errorf("Membership.PermitStripeUpgrade() = %v, want %v", got, tt.want)
			}
		})
	}
}
