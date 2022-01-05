package price

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/guregu/null"
	"testing"
)

func TestFloat(t *testing.T) {
	diff := 35.00 - 34.99

	t.Logf("%f", diff)
}

func TestNewCharge(t *testing.T) {
	charge := NewCharge(Price{
		ID:         "",
		Edition:    Edition{},
		Active:     false,
		Archived:   false,
		Currency:   "cny",
		Title:      null.String{},
		LiveMode:   false,
		Nickname:   null.String{},
		ProductID:  "",
		UnitAmount: 35,
		CreatedUTC: chrono.Time{},
	}, Discount{
		ID: "",
		DiscountParams: DiscountParams{
			CreatedBy:    "",
			Description:  null.String{},
			Kind:         "",
			Percent:      null.Int{},
			ChronoPeriod: dt.ChronoPeriod{},
			PriceOff:     null.FloatFrom(34.99),
			PriceID:      "",
			Recurring:    false,
		},
		LiveMode:   false,
		Status:     "",
		CreatedUTC: chrono.Time{},
	})

	t.Logf("%+v", charge)
}
