package stripe

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/addon"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/invoice"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"reflect"
	"testing"
	"time"
)

func TestWebhookResultBuilder_Build(t *testing.T) {
	ftcID := uuid.New().String()

	userIDs := ids.UserIDs{
		CompoundID: ftcID,
		FtcID:      null.StringFrom(ftcID),
		UnionID:    null.String{},
	}

	ftcMmb := reader.Membership{
		UserIDs:       userIDs,
		Edition:       price.StdYearEdition,
		LegacyTier:    null.Int{},
		LegacyExpire:  null.Int{},
		ExpireDate:    chrono.DateFrom(time.Now().AddDate(0, 0, 30)),
		PaymentMethod: enum.PayMethodAli,
	}

	addOn := ftcMmb.CarriedOverAddOn()

	subs := NewMockSubsBuilder(ftcID).Build()

	type fields struct {
		Subs         Subs
		UserIDs      ids.UserIDs
		StripeMember reader.Membership
		FtcMember    reader.Membership
	}
	tests := []struct {
		name    string
		fields  fields
		want    WebhookResult
		wantErr bool
	}{
		{
			name: "Both side empty",
			fields: fields{
				Subs:         subs,
				UserIDs:      userIDs,
				StripeMember: reader.Membership{},
				FtcMember:    reader.Membership{},
			},
			want: WebhookResult{
				Member: NewMembership(MembershipParams{
					UserIDs: userIDs,
					Subs:    subs,
					AddOn:   addon.AddOn{},
				}),
				Versioned:        reader.MembershipVersioned{},
				CarryOverInvoice: invoice.Invoice{},
			},
			wantErr: false,
		},
		{
			name: "Ftc side expired",
			fields: fields{
				Subs:         subs,
				UserIDs:      userIDs,
				StripeMember: reader.Membership{},
				FtcMember: reader.Membership{
					UserIDs:       userIDs,
					Edition:       price.StdYearEdition,
					LegacyTier:    null.Int{},
					LegacyExpire:  null.Int{},
					ExpireDate:    chrono.DateFrom(time.Now().AddDate(0, -1, 0)),
					PaymentMethod: enum.PayMethodAli,
				},
			},
			want: WebhookResult{
				Member: NewMembership(MembershipParams{
					UserIDs: userIDs,
					Subs:    subs,
					AddOn:   addon.AddOn{},
				}),
				Versioned:        reader.MembershipVersioned{},
				CarryOverInvoice: invoice.Invoice{},
			},
			wantErr: false,
		},
		{
			name: "Ftc side valid while stripe expired",
			fields: fields{
				Subs:         NewMockSubsBuilder(ftcID).WithCanceled().Build(),
				UserIDs:      userIDs,
				StripeMember: reader.Membership{},
				FtcMember: reader.Membership{
					UserIDs:       userIDs,
					Edition:       price.StdYearEdition,
					LegacyTier:    null.Int{},
					LegacyExpire:  null.Int{},
					ExpireDate:    chrono.DateFrom(time.Now().AddDate(0, 0, 30)),
					PaymentMethod: enum.PayMethodAli,
				},
			},
			want:    WebhookResult{},
			wantErr: true,
		},
		{
			name: "Valid stripe overrides one-time",
			fields: fields{
				Subs:         subs,
				UserIDs:      userIDs,
				StripeMember: reader.Membership{},
				FtcMember:    ftcMmb,
			},
			want: WebhookResult{
				Member: NewMembership(MembershipParams{
					UserIDs: userIDs,
					Subs:    subs,
					AddOn:   addOn,
				}),
				Versioned:        reader.MembershipVersioned{},
				CarryOverInvoice: ftcMmb.CarryOverInvoice(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := WebhookResultBuilder{
				Subs:         tt.fields.Subs,
				UserIDs:      tt.fields.UserIDs,
				StripeMember: tt.fields.StripeMember,
				FtcMember:    tt.fields.FtcMember,
			}
			got, err := b.Build()
			if (err != nil) != tt.wantErr {
				t.Errorf("Build() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.Member, tt.want.Member) {
				t.Errorf("Build() got = %v, want %v", got, tt.want)
			}

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}
