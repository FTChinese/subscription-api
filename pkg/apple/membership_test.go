package apple

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg"
	"github.com/FTChinese/subscription-api/pkg/addon"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"reflect"
	"testing"
	"time"
)

func TestNewMembership(t *testing.T) {
	userID := uuid.New().String()

	now := time.Now()
	txID := faker.GenAppleSubID()

	type args struct {
		params MembershipParams
	}
	tests := []struct {
		name string
		args args
		want reader.Membership
	}{
		{
			name: "Build membership",
			args: args{params: MembershipParams{
				UserID: pkg.NewFtcUserID(userID),
				Subs: NewMockSubsBuilder(userID).
					WithOriginalTxID(txID).
					Build(),
				AddOn: addon.AddOn{},
			}},
			want: reader.Membership{
				UserIDs:       pkg.NewFtcUserID(userID),
				Edition:       price.StdYearEdition,
				LegacyTier:    null.Int{},
				LegacyExpire:  null.Int{},
				ExpireDate:    chrono.DateFrom(now.AddDate(1, 0, 0)),
				PaymentMethod: enum.PayMethodApple,
				FtcPlanID:     null.String{},
				StripeSubsID:  null.String{},
				StripePlanID:  null.String{},
				AutoRenewal:   true,
				Status:        0,
				AppleSubsID:   null.StringFrom(txID),
				B2BLicenceID:  null.String{},
				AddOn:         addon.AddOn{},
			}.Sync(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewMembership(tt.args.params); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMembership() = %v, want %v", got, tt.want)
			}
		})
	}
}
