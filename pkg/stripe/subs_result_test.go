package stripe

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
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
					UserIDs:       member.UserIDs,
					Kind:          reader.SubsKindOneTimeToSub,
					CurrentMember: member,
					Action:        reader.ActionActionCreate,
				},
			},
			want: SubsResult{
				Modified:             true,
				MissingPaymentIntent: false,
				Subs:                 subs,
				Member: reader.Membership{
					UserIDs:       member.UserIDs,
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
					AddOn:         member.CarriedOverAddOn(),
				}.Sync(),
				Snapshot: member.Snapshot(reader.NewStripeArchiver(reader.ActionActionCreate)),
				CarryOverInvoice: member.CarryOverInvoice().
					WithStripeSubsID(subs.ID),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newSubsResult(tt.args.subs, tt.args.params)

			tt.want.Subs.UpdatedUTC = got.Subs.UpdatedUTC
			tt.want.Snapshot.CreatedUTC = got.Snapshot.CreatedUTC
			tt.want.CarryOverInvoice.ID = got.CarryOverInvoice.ID
			tt.want.CarryOverInvoice.CreatedUTC = got.CarryOverInvoice.CreatedUTC
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

			if !reflect.DeepEqual(got.CarryOverInvoice, tt.want.CarryOverInvoice) {
				t.Errorf("newSubsResult().AddOn = \n%+v, \nwant %+v", got.CarryOverInvoice, tt.want.CarryOverInvoice)
			}
		})
	}
}
