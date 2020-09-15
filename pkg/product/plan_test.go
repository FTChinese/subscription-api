package product

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
	"github.com/stretchr/testify/assert"
	"reflect"
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

func TestExpandedPlan_Payable(t *testing.T) {

	tests := []struct {
		name   string
		fields ExpandedPlan
		want   Charge
	}{
		{
			name:   "With discount",
			fields: planStdYear,
			want: Charge{
				Amount:     128,
				DiscountID: null.StringFrom("dsc_F7gEwjaF3OsR"),
				Currency:   "cny",
			},
		},
		{
			name:   "No discount",
			fields: planStdMonth,
			want: Charge{
				Amount:     28,
				DiscountID: null.String{},
				Currency:   "cny",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := ExpandedPlan{
				Plan:     tt.fields.Plan,
				Discount: tt.fields.Discount,
			}
			if got := e.Payable(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Payable() = %v, want %v", got, tt.want)
			}
		})
	}
}
