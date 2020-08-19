package product

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
	"github.com/stretchr/testify/assert"
	"testing"
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
		Description: null.String{},
	},
	Discount: Discount{
		DiscID:   null.StringFrom("dsc_F7gEwjaF3OsR"),
		PriceOff: null.FloatFrom(130),
		Percent:  null.Int{},
		Period: Period{
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
		Description: null.String{},
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
		Description: null.String{},
	},
	Discount: Discount{},
}

func TestExpandedPlan_Amount(t *testing.T) {
	type fields struct {
		Plan ExpandedPlan
	}
	tests := []struct {
		name   string
		fields fields
		want   float64
	}{
		{
			name: "Standard yearly plan",
			fields: fields{
				Plan: planStdYear,
			},
			want: 128,
		},
		{
			name:   "Standard monthly plan",
			fields: fields{Plan: planStdMonth},
			want:   28,
		},
		{
			name:   "Premium plan",
			fields: fields{Plan: planPrm},
			want:   1998,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := tt.fields.Plan

			got := e.Amount()

			assert.Equal(t, got, tt.want)
		})
	}
}
