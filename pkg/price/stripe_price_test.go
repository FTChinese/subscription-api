package price

import (
	"testing"
)

func TestStripePrice_uniqueFeatures(t *testing.T) {

	tests := []struct {
		name   string
		fields StripePrice
		want   string
	}{
		{
			name:   "Intro price",
			fields: MockStripeStdIntroPrice,
			want:   "stripe.standard.null.one_time.sandbox",
		},
		{
			name:   "Standard year recurring",
			fields: MockStripeStdYearPrice,
			want:   "stripe.standard.year.recurring.sandbox",
		},
		{
			name:   "Standard month recurring",
			fields: MockStripeStdMonthPrice,
			want:   "stripe.standard.month.recurring.sandbox",
		},
		{
			name:   "Standard premium recurring",
			fields: MockStripePrmPrice,
			want:   "stripe.premium.year.recurring.sandbox",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.fields
			if got := p.uniqueFeatures(); got != tt.want {
				t.Errorf("StripePrice.uniqueFeatures() = %v, want %v", got, tt.want)
			}
		})
	}
}
