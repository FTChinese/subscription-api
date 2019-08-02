package paywall

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
	"testing"
)

func TestStripeSubParams_SetStripePlanID(t *testing.T) {
	type fields struct {
		Coordinate           Coordinate
		Customer             string
		Coupon               null.String
		DefaultPaymentMethod null.String
		IdempotencyKey       string
		planID               string
	}
	type args struct {
		live bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Standard Month",
			fields: fields{
				Coordinate: Coordinate{
					Tier:  enum.TierStandard,
					Cycle: enum.CycleMonth,
				},
			},
			args:    args{live: false},
			wantErr: false,
		},
		{
			name: "Standard Year",
			fields: fields{
				Coordinate: Coordinate{
					Tier:  enum.TierStandard,
					Cycle: enum.CycleYear,
				},
			},
			args:    args{live: false},
			wantErr: false,
		},
		{
			name: "Premium Year",
			fields: fields{
				Coordinate: Coordinate{
					Tier:  enum.TierPremium,
					Cycle: enum.CycleYear,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &StripeSubParams{
				Coordinate:           tt.fields.Coordinate,
				Customer:             tt.fields.Customer,
				Coupon:               tt.fields.Coupon,
				DefaultPaymentMethod: tt.fields.DefaultPaymentMethod,
				IdempotencyKey:       tt.fields.IdempotencyKey,
				planID:               tt.fields.planID,
			}
			if err := p.SetStripePlanID(tt.args.live); (err != nil) != tt.wantErr {
				t.Errorf("SetStripePlanID() error = %v, wantErr %v", err, tt.wantErr)
			}

			t.Logf("%s: %s", p.NamedKey(), p.planID)
		})
	}
}
