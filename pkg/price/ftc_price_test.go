package price

import (
	"testing"
)

func TestFtcPrice_uniqueFeatures(t *testing.T) {

	tests := []struct {
		name   string
		fields FtcPrice
		want   string
	}{
		{
			name:   "standard year intro",
			fields: MockFtcStdIntroPrice,
			want:   "ftc.standard..one_time.sandbox",
		},
		{
			name:   "standard year",
			fields: MockFtcStdYearPrice,
			want:   "ftc.standard.year.recurring.sandbox",
		},
		{
			name:   "standard month",
			fields: MockFtcStdMonthPrice,
			want:   "ftc.standard.month.recurring.sandbox",
		},
		{
			name:   "premium year",
			fields: MockFtcPrmPrice,
			want:   "ftc.premium.year.recurring.sandbox",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := FtcPrice{
				ID:            tt.fields.ID,
				Edition:       tt.fields.Edition,
				Active:        tt.fields.Active,
				Archived:      tt.fields.Archived,
				Currency:      tt.fields.Currency,
				Kind:          tt.fields.Kind,
				LiveMode:      tt.fields.LiveMode,
				Nickname:      tt.fields.Nickname,
				PeriodCount:   tt.fields.PeriodCount,
				ProductID:     tt.fields.ProductID,
				StripePriceID: tt.fields.StripePriceID,
				Title:         tt.fields.Title,
				UnitAmount:    tt.fields.UnitAmount,
				StartUTC:      tt.fields.StartUTC,
				EndUTC:        tt.fields.EndUTC,
				CreatedUTC:    tt.fields.CreatedUTC,
			}
			if got := p.uniqueFeatures(); got != tt.want {
				t.Errorf("FtcPrice.uniqueFeatures() = %v, want %v", got, tt.want)
			}
		})
	}
}
