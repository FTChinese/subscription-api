package subs

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/product"
	"github.com/guregu/null"
	"time"
)

var planStdYear = product.ExpandedPlan{
	Plan: product.Plan{
		ID:        "plan_MynUQDQY1TSQ",
		ProductID: "prod_zjWdiTUpDN8l",
		Price:     258,
		Edition: product.Edition{
			Tier:  enum.TierStandard,
			Cycle: enum.CycleYear,
		},
		Description: null.String{},
	},
	Discount: product.Discount{
		DiscID:   null.StringFrom("dsc_F7gEwjaF3OsR"),
		PriceOff: null.FloatFrom(130),
		Percent:  null.Int{},
		Period: product.Period{
			StartUTC: chrono.TimeNow(),
			EndUTC:   chrono.TimeFrom(time.Now().AddDate(0, 0, 2)),
		},
		Description: null.String{},
	},
}
