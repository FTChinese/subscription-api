package paywall

import (
	"github.com/FTChinese/go-rest/enum"
	"testing"
)

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
