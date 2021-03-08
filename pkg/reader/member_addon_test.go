package reader

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/addon"
	"github.com/FTChinese/subscription-api/pkg/invoice"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"reflect"
	"testing"
	"time"
)

func TestMembership_claimAddOn(t *testing.T) {

	userID := uuid.New().String()

	inv := invoice.NewMockInvoiceBuilder(userID).
		WithOrderKind(enum.OrderKindAddOn).
		Build()

	current := NewMockMemberBuilder(userID).
		WithExpiration(time.Now().AddDate(0, 0, -1)).
		Build().
		PlusAddOn(addon.New(inv.Tier, inv.TotalDays()))

	type args struct {
		i invoice.Invoice
	}
	tests := []struct {
		name    string
		fields  Membership
		args    args
		want    Membership
		wantErr bool
	}{
		{
			name:   "Transfer add on",
			fields: current,
			args: args{
				i: inv,
			},
			want: Membership{
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

			got, err := tt.fields.claimAddOn(tt.args.i)
			if (err != nil) != tt.wantErr {
				t.Errorf("claimAddOn() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("claimAddOn() got = %v, want %v", got, tt.want)
			}
		})
	}
}
