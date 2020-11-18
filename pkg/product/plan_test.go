package product

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/dt"
	"github.com/guregu/null"
	"time"
)

var planStdYear = ExpandedPlan{
	Plan: Plan{
		ID:        "plan_MynUQDQY1TSQ",
		ProductID: "prod_zjWdiTUpDN8l",
		Price:     258,
		Edition: Edition{
			Tier:  enum.TierStandard,
			Cycle: enum.CycleYear,
		},
		Description: "",
	},
	Discount: Discount{
		DiscID:   null.StringFrom("dsc_F7gEwjaF3OsR"),
		PriceOff: null.FloatFrom(130),
		Percent:  null.Int{},
		DateTimeRange: dt.DateTimeRange{
			StartUTC: chrono.TimeNow(),
			EndUTC:   chrono.TimeFrom(time.Now().AddDate(0, 0, 2)),
		},
		Description: null.String{},
	},
}

var planStdMonth = ExpandedPlan{
	Plan: Plan{
		ID:        "plan_1Uz4hrLy3Mzy",
		ProductID: "prod_zjWdiTUpDN8l",
		Price:     28,
		Edition: Edition{
			Tier:  enum.TierStandard,
			Cycle: enum.CycleMonth,
		},
		Description: "",
	},
	Discount: Discount{},
}

var planPrm = ExpandedPlan{
	Plan: Plan{
		ID:        "plan_vRUzRQ3aglea",
		ProductID: "prod_IaoK5SbK79g8",
		Price:     1998,
		Edition: Edition{
			Tier:  enum.TierPremium,
			Cycle: enum.CycleYear,
		},
		Description: "",
	},
	Discount: Discount{},
}
