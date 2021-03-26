package price

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/guregu/null"
	"testing"
	"time"
)

func TestDiscount_IsValid(t *testing.T) {

	tests := []struct {
		name   string
		fields Discount
		want   bool
	}{
		{
			name:   "Zero value",
			fields: Discount{},
			want:   false,
		},
		{
			name: "Retention offer",
			fields: Discount{
				DiscID:         null.StringFrom("random-id"),
				PriceOff:       null.FloatFrom(10),
				Percent:        null.IntFrom(90),
				DateTimePeriod: dt.DateTimePeriod{},
				Description:    null.String{},
				Kind:           OfferKindRetention,
			},
			want: true,
		},
		{
			name: "Promotion offer",
			fields: Discount{
				DiscID:   null.StringFrom("random-id"),
				PriceOff: null.FloatFrom(10),
				Percent:  null.IntFrom(90),
				DateTimePeriod: dt.DateTimePeriod{
					StartUTC: chrono.TimeNow(),
					EndUTC:   chrono.TimeFrom(time.Now().AddDate(1, 0, 0)),
				},
				Description: null.String{},
				Kind:        OfferKindPromotion,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := Discount{
				DiscID:         tt.fields.DiscID,
				PriceOff:       tt.fields.PriceOff,
				Percent:        tt.fields.Percent,
				DateTimePeriod: tt.fields.DateTimePeriod,
				Description:    tt.fields.Description,
				Kind:           tt.fields.Kind,
			}
			if got := d.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}
