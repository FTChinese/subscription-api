package price

import (
	"testing"
	"time"
)

func TestStripeCoupon_IsValid(t *testing.T) {

	tests := []struct {
		name   string
		fields StripeCoupon
		want   bool
	}{
		{
			name: "",
			fields: StripeCoupon{
				RedeemBy: time.Now().AddDate(0, 0, 7).Unix(),
				Status:   DiscountStatusActive,
			},
			want: true,
		},

		{
			name: "",
			fields: StripeCoupon{
				RedeemBy: time.Now().AddDate(0, 0, -7).Unix(),
				Status:   DiscountStatusActive,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.fields

			if got := c.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}
