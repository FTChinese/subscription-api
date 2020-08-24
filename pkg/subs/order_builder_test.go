package subs

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/product"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"github.com/stretchr/testify/assert"
	"testing"
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

func TestOrderBuilder_AliAppPayParams(t *testing.T) {
	builder := NewOrderBuilder(reader.MemberID{
		CompoundID: "",
		FtcID:      null.StringFrom(uuid.New().String()),
		UnionID:    null.String{},
	}.MustNormalize()).
		SetPlan(planStdYear).
		SetPayMethod(enum.PayMethodAli).
		SetEnvironment(true)

	err := builder.DeduceSubsKind(reader.Membership{})
	if err != nil {
		t.Error(err)
	}

	err = builder.Build()

	if err != nil {
		t.Error(err)
	}

	order, err := builder.GetOrder()

	t.Logf("%+v", order)

	aliPayIntent := builder.AliAppPayParams()

	assert.NotEqual(t, aliPayIntent.TotalAmount, 0.01)

	t.Logf("%+v", aliPayIntent)
}
