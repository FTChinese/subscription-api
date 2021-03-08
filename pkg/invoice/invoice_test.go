package invoice

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/addon"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"reflect"
	"testing"
	"time"
)

func TestInvoice_NewMembership(t *testing.T) {
	userID := uuid.New().String()

	current := reader.NewMockMemberBuilder(userID).Build()

	type args struct {
		userID  reader.MemberID
		current reader.Membership
	}
	tests := []struct {
		name    string
		fields  Invoice
		args    args
		want    reader.Membership
		wantErr bool
	}{
		{
			name:   "Create membership",
			fields: NewMockInvoiceBuilder(userID).SetPeriodStart(time.Now()).Build(),
			args: args{
				userID:  reader.NewFtcUserID(userID),
				current: reader.Membership{},
			},
			want: reader.Membership{
				MemberID:      reader.NewFtcUserID(userID),
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
			name: "Renew membership",
			fields: NewMockInvoiceBuilder(userID).
				WithOrderKind(enum.OrderKindRenew).
				SetPeriodStart(current.ExpireDate.Time).
				Build(),
			args: args{
				userID:  reader.NewFtcUserID(userID),
				current: current,
			},
			want: reader.Membership{
				MemberID:      reader.NewFtcUserID(userID),
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
			name: "Upgrade membership",
			fields: NewMockInvoiceBuilder(userID).
				WithPrice(faker.PricePrm).
				WithOrderKind(enum.OrderKindUpgrade).
				SetPeriodStart(time.Now()).
				Build(),
			args: args{
				userID:  reader.NewFtcUserID(userID),
				current: current,
			},
			want: reader.Membership{
				MemberID:      reader.NewFtcUserID(userID),
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
			name: "Membership addon",
			fields: NewMockInvoiceBuilder(userID).
				WithOrderKind(enum.OrderKindAddOn).
				Build(),
			args: args{
				userID:  reader.NewFtcUserID(userID),
				current: current,
			},
			want: current.WithAddOn(addon.AddOn{
				Standard: 367,
				Premium:  0,
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.fields.NewMembership(tt.args.userID, tt.args.current)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewMembership() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMembership() = \n%v, want \n%v", got, tt.want)
			}
		})
	}
}

func TestInvoice_TransferAddOn(t *testing.T) {
	userID := uuid.New().String()

	inv := NewMockInvoiceBuilder(userID).
		WithOrderKind(enum.OrderKindAddOn).
		Build()

	current := reader.NewMockMemberBuilder(userID).
		WithExpiration(time.Now().AddDate(0, 0, -1)).
		Build().
		WithAddOn(addon.New(inv.Tier, inv.TotalDays()))

	type args struct {
		current reader.Membership
	}
	tests := []struct {
		name    string
		fields  Invoice
		args    args
		want    reader.Membership
		wantErr bool
	}{
		{
			name:   "Transfer add on",
			fields: NewMockInvoiceBuilder(userID).WithOrderKind(enum.OrderKindAddOn).Build(),
			args: args{
				current: current,
			},
			want: reader.Membership{
				MemberID:      current.MemberID,
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
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := tt.fields.TransferAddOn(tt.args.current)
			if (err != nil) != tt.wantErr {
				t.Errorf("TransferAddOn() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TransferAddOn() got = %v, want %v", got, tt.want)
			}
		})
	}
}
