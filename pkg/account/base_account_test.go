package account

import (
	"github.com/guregu/null"
	"reflect"
	"testing"
)

func TestBaseAccount_IsMobileEmail(t *testing.T) {

	tests := []struct {
		name   string
		fields BaseAccount
		want   bool
	}{
		{
			name: "Mobile email",
			fields: BaseAccount{
				Email:  "1234567890" + mobileEmailSuffix,
				Mobile: null.String{},
			},
			want: true,
		},
		{
			name: "Not mobile email",
			fields: BaseAccount{
				Email:  "1234567890" + mobileEmailSuffix,
				Mobile: null.StringFrom("1234567890"),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := BaseAccount{
				FtcID:        tt.fields.FtcID,
				UnionID:      tt.fields.UnionID,
				StripeID:     tt.fields.StripeID,
				Email:        tt.fields.Email,
				Password:     tt.fields.Password,
				Mobile:       tt.fields.Mobile,
				UserName:     tt.fields.UserName,
				AvatarURL:    tt.fields.AvatarURL,
				IsVerified:   tt.fields.IsVerified,
				CampaignCode: tt.fields.CampaignCode,
			}
			if got := a.IsMobileEmail(); got != tt.want {
				t.Errorf("IsMobileEmail() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBaseAccount_SyncMobile(t *testing.T) {

	tests := []struct {
		name   string
		fields BaseAccount
		want   BaseAccount
	}{
		{
			name: "Mobile email",
			fields: BaseAccount{
				Email:  "13945678900" + mobileEmailSuffix,
				Mobile: null.String{},
			},
			want: BaseAccount{
				Email:  "13945678900" + mobileEmailSuffix,
				Mobile: null.StringFrom("13945678900"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := BaseAccount{
				FtcID:        tt.fields.FtcID,
				UnionID:      tt.fields.UnionID,
				StripeID:     tt.fields.StripeID,
				Email:        tt.fields.Email,
				Password:     tt.fields.Password,
				Mobile:       tt.fields.Mobile,
				UserName:     tt.fields.UserName,
				AvatarURL:    tt.fields.AvatarURL,
				IsVerified:   tt.fields.IsVerified,
				CampaignCode: tt.fields.CampaignCode,
			}
			if got := a.SyncMobile(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UpsertMobile() = %v, want %v", got, tt.want)
			}
		})
	}
}
