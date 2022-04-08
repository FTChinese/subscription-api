package stripe

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/addon"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"reflect"
	"testing"
	"time"
)

func TestNewMembership(t *testing.T) {

	ftcID := uuid.New().String()

	type args struct {
		params MembershipParams
	}
	tests := []struct {
		name string
		args args
		want reader.Membership
	}{
		{
			name: "New membership",
			args: args{
				params: MembershipParams{
					UserIDs: ids.UserIDs{
						CompoundID: ftcID,
						FtcID:      null.StringFrom(ftcID),
					},
					Subs: Subs{
						IsFromStripe:           false,
						ID:                     faker.GenStripeSubID(),
						Edition:                price.MockEditionStdYear,
						WillCancelAtUtc:        chrono.Time{},
						CancelAtPeriodEnd:      false,
						CanceledUTC:            chrono.Time{},
						CurrentPeriodEnd:       chrono.TimeFrom(time.Now().AddDate(1, 0, 0)),
						CurrentPeriodStart:     chrono.TimeFrom(time.Now()),
						CustomerID:             faker.GenStripeCusID(),
						DefaultPaymentMethodID: null.StringFrom(faker.GenPaymentMethodID()),
						EndedUTC:               chrono.Time{},
						FtcUserID:              null.StringFrom(ftcID),
						Items: []SubsItem{
							{
								ID: faker.GenStripeItemID(),
								Price: PriceColumn{
									StripePrice: MockPriceStdYear,
								},
								Created:        time.Now().Unix(),
								Quantity:       1,
								SubscriptionID: "",
							},
						},
						LatestInvoiceID: "",
						LatestInvoice:   Invoice{},
						LiveMode:        false,
						PaymentIntentID: null.StringFrom(faker.GenPaymentIntentID()),
						PaymentIntent:   PaymentIntent{},
						StartDateUTC:    chrono.TimeNow(),
						Status:          enum.SubsStatusActive,
						Created:         time.Now().Unix(),
						ItemID:          "",
					},
					AddOn: addon.AddOn{},
				},
			},
			want: reader.Membership{
				UserIDs: ids.UserIDs{
					CompoundID: ftcID,
					FtcID:      null.StringFrom(ftcID),
				},
				Edition:       price.StdYearEdition,
				LegacyTier:    null.IntFrom(reader.GetTierCode(enum.TierStandard)),
				LegacyExpire:  null.IntFrom(1638943057),
				ExpireDate:    chrono.DateFrom(time.Unix(1638943057, 0)),
				PaymentMethod: enum.PayMethodStripe,
				FtcPlanID:     null.String{},
				StripeSubsID:  null.StringFrom("sub_IX3JAkik1JKDzW"),
				StripePlanID:  null.StringFrom(MockPriceStdYear.ID),
				AutoRenewal:   true,
				Status:        enum.SubsStatusActive,
				AppleSubsID:   null.String{},
				B2BLicenceID:  null.String{},
				AddOn:         addon.AddOn{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewMembership(tt.args.params); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMembership() = %v\n, want %v", got, tt.want)
			}
		})
	}
}
