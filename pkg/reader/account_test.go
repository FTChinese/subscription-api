package reader

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/addon"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"reflect"
	"testing"
	"time"
)

func TestAccount_VerifyDelete(t *testing.T) {
	type fields struct {
		BaseAccount account.BaseAccount
		LoginMethod enum.LoginMethod
		Wechat      account.Wechat
		Membership  Membership
	}
	type args struct {
		email string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *render.ValidationError
	}{
		{
			name: "Email does not match",
			fields: fields{
				BaseAccount: account.BaseAccount{
					FtcID:        uuid.New().String(),
					UnionID:      null.String{},
					StripeID:     null.String{},
					Email:        "test@example.org",
					Password:     "",
					Mobile:       null.String{},
					UserName:     null.String{},
					AvatarURL:    null.String{},
					IsVerified:   false,
					CampaignCode: null.String{},
				},
				LoginMethod: 0,
				Wechat:      account.Wechat{},
				Membership:  Membership{},
			},
			args: args{
				email: "test1@example.org",
			},
			want: &render.ValidationError{
				Message: "Email mismatched",
				Field:   "email",
				Code:    "missing",
			},
		},
		{
			name: "Membership not expired",
			fields: fields{
				BaseAccount: account.BaseAccount{
					FtcID:        uuid.New().String(),
					UnionID:      null.String{},
					StripeID:     null.String{},
					Email:        "test@example.org",
					Password:     "",
					Mobile:       null.String{},
					UserName:     null.String{},
					AvatarURL:    null.String{},
					IsVerified:   false,
					CampaignCode: null.String{},
				},
				LoginMethod: 0,
				Wechat:      account.Wechat{},
				Membership: Membership{
					UserIDs: ids.UserIDs{
						CompoundID: uuid.New().String(),
						FtcID:      null.StringFrom(uuid.New().String()),
						UnionID:    null.String{},
					},
					Edition: price.Edition{
						Tier:  enum.TierStandard,
						Cycle: enum.CycleYear,
					},
					LegacyTier:    null.Int{},
					LegacyExpire:  null.Int{},
					ExpireDate:    chrono.DateFrom(time.Now().AddDate(1, 0, 0)),
					PaymentMethod: 0,
					FtcPlanID:     null.String{},
					StripeSubsID:  null.String{},
					StripePlanID:  null.String{},
					AutoRenewal:   false,
					Status:        0,
					AppleSubsID:   null.String{},
					B2BLicenceID:  null.String{},
					AddOn:         addon.AddOn{},
					VIP:           false,
				}.Sync(),
			},
			args: args{
				email: "test@example.org",
			},
			want: &render.ValidationError{
				Message: "Subscription is still valid",
				Field:   "subscription",
				Code:    "already_exists",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := Account{
				BaseAccount: tt.fields.BaseAccount,
				LoginMethod: tt.fields.LoginMethod,
				Wechat:      tt.fields.Wechat,
				Membership:  tt.fields.Membership,
			}
			if got := a.VerifyDelete(tt.args.email); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("VerifyDelete() = %v, want %v", got, tt.want)
			}
		})
	}
}
