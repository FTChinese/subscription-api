package model

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"github.com/stripe/stripe-go"
	"gitlab.com/ftchinese/subscription-api/test"
	"testing"

	"gitlab.com/ftchinese/subscription-api/paywall"
)

func Test_createStripeSub(t *testing.T) {
	stripe.Key = test.StripeKey

	type args struct {
		p paywall.StripeSubParams
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Create Stripe Subscription",
			args: args{
				p: paywall.StripeSubParams{
					Tier:                 enum.TierStandard,
					Cycle:                enum.CycleYear,
					Customer:             "cus_FOgRRgj9aMzpAv",
					DefaultPaymentMethod: null.StringFrom("pm_1Ett5HBzTK0hABgJwXpA8b7z"),
					PlanID:               "plan_FOdfeaqzczp6Ag",
					IdempotencyKey:       uuid.New().String(),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := createStripeSub(tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("createStripeSub() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("Stripe subscripiton: %+v", got)
		})
	}
}
