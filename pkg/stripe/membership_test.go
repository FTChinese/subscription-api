package stripe

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/addon"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/guregu/null"
	"reflect"
	"testing"
	"time"
)

func TestNewMembership(t *testing.T) {
	subs := MockNewSubs()

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
					UserIDs: reader.MemberID{
						CompoundID: subs.FtcUserID.String,
						FtcID:      subs.FtcUserID,
					},
					Subs:         subs,
					ReservedDays: addon.AddOn{},
				},
			},
			want: reader.Membership{
				MemberID: reader.MemberID{
					CompoundID: subs.FtcUserID.String,
					FtcID:      subs.FtcUserID,
				},
				Edition:       price.PremiumEdition,
				LegacyTier:    null.IntFrom(reader.GetTierCode(enum.TierPremium)),
				LegacyExpire:  null.IntFrom(1638943057),
				ExpireDate:    chrono.DateFrom(time.Unix(1638943057, 0)),
				PaymentMethod: enum.PayMethodStripe,
				FtcPlanID:     null.String{},
				StripeSubsID:  null.StringFrom("sub_IX3JAkik1JKDzW"),
				StripePlanID:  null.StringFrom("plan_FOde0uAr0V4WmT"),
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
