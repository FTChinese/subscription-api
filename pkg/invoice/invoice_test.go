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
		name   string
		fields Invoice
		args   args
		want   reader.Membership
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

			if got := tt.fields.NewMembership(tt.args.userID, tt.args.current); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMembership() = \n%v, want \n%v", got, tt.want)
			}
		})
	}
}
