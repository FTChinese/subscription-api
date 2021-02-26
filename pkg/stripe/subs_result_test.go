package stripe

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/addon"
	"github.com/FTChinese/subscription-api/pkg/cart"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"reflect"
	"testing"
)

func Test_newSubsResult(t *testing.T) {
	userID := uuid.New().String()
	member := reader.NewMockMemberBuilder(userID).Build()
	subs := NewMockSubsBuilder(userID).Build()

	type args struct {
		subs   Subs
		params SubsResultParams
	}
	tests := []struct {
		name string
		args args
		want SubsResult
	}{
		{
			name: "New subscription switching form one-time",
			args: args{
				subs: subs,
				params: SubsResultParams{
					UserIDs:       member.MemberID,
					Kind:          cart.SubsKindOneTimeToStripe,
					CurrentMember: member,
					Action:        reader.ActionCreate,
				},
			},
			want: SubsResult{
				Modified:             true,
				MissingPaymentIntent: false,
				Subs:                 subs,
				Member: reader.Membership{
					MemberID:      member.MemberID,
					Edition:       subs.Edition,
					LegacyTier:    null.IntFrom(reader.GetTierCode(subs.Tier)),
					LegacyExpire:  null.IntFrom(subs.ExpiresAt().Unix()),
					ExpireDate:    chrono.DateFrom(subs.ExpiresAt()),
					PaymentMethod: enum.PayMethodStripe,
					FtcPlanID:     null.String{},
					StripeSubsID:  null.StringFrom(subs.ID),
					StripePlanID:  null.StringFrom(subs.PriceID),
					AutoRenewal:   true,
					Status:        enum.SubsStatusActive,
					AppleSubsID:   null.String{},
					B2BLicenceID:  null.String{},
					ReservedDays:  member.CarryOver(addon.CarryOverFromSwitchingStripe).ToReservedDays(),
				}.Sync(),
				Snapshot: member.Snapshot(reader.StripeArchiver(reader.ActionCreate)),
				AddOn:    member.CarryOver(addon.CarryOverFromSwitchingStripe),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newSubsResult(tt.args.subs, tt.args.params)

			tt.want.Subs.UpdatedUTC = got.Subs.UpdatedUTC
			tt.want.Snapshot.CreatedUTC = got.Snapshot.CreatedUTC
			tt.want.AddOn.ID = got.AddOn.ID
			tt.want.AddOn.CreatedUTC = got.AddOn.CreatedUTC
			tt.want.Snapshot.SnapshotID = got.Snapshot.SnapshotID
			tt.want.Snapshot.CreatedUTC = got.Snapshot.CreatedUTC

			if !reflect.DeepEqual(got.Subs, tt.want.Subs) {
				t.Errorf("newSubsResult().Subs =\n%+v, \nwant \n%+v", got.Subs, tt.want.Subs)
			}

			if !reflect.DeepEqual(got.Member, tt.want.Member) {
				t.Errorf("newSubsResult().Member = \n%+v, \nwant \n%+v", got.Member, tt.want.Member)
			}

			if !reflect.DeepEqual(got.Snapshot, tt.want.Snapshot) {
				t.Errorf("newSubsResult().Snapshot = \n%+v, \nwant \n%+v", got.Snapshot, tt.want.Snapshot)
			}

			if !reflect.DeepEqual(got.AddOn, tt.want.AddOn) {
				t.Errorf("newSubsResult().AddOn = \n%+v, \nwant %+v", got.AddOn, tt.want.AddOn)
			}
		})
	}
}
