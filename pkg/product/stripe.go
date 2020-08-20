package product

import "github.com/FTChinese/go-rest/enum"

type NonFtcPlan struct {
	Edition
	stripeLivePlanID string
	stripeTestPlanID string
	AppleProductID   string
}

func (s NonFtcPlan) GetID(live bool) string {
	if live {
		return s.stripeLivePlanID
	}

	return s.stripeTestPlanID
}

var stdMonth = NonFtcPlan{
	Edition: Edition{
		Tier:  enum.TierStandard,
		Cycle: enum.CycleMonth,
	},
	stripeLivePlanID: "plan_FXZYLOEbcvj5Tx",
	stripeTestPlanID: "plan_FOdgPTznDwHU4i",
	AppleProductID:   "com.ft.ftchinese.mobile.subscription.member.monthly",
}

var stdYear = NonFtcPlan{
	Edition: Edition{
		Tier:  enum.TierStandard,
		Cycle: enum.CycleYear,
	},
	stripeLivePlanID: "plan_FXZZUEDpToPlZK",
	stripeTestPlanID: "plan_FOdfeaqzczp6Ag",
	AppleProductID:   "com.ft.ftchinese.mobile.subscription.member",
}

var prmYear = NonFtcPlan{
	Edition:          NewPremiumEdition(),
	stripeLivePlanID: "plan_FXZbv1cDTsUKOg",
	stripeTestPlanID: "plan_FOde0uAr0V4WmT",
	AppleProductID:   "com.ft.ftchinese.mobile.subscription.vip",
}

type NonFtcPlanStore struct {
	plans               []NonFtcPlan
	indexStripeLivePlan map[string]int
	indexStripeTestPlan map[string]int
	indexAppleProduct   map[string]int
}

var NonFtcPlans = NonFtcPlanStore{
	plans: []NonFtcPlan{
		stdMonth,
		stdYear,
		prmYear,
	},
	indexStripeLivePlan: map[string]int{},
	indexStripeTestPlan: map[string]int{},
	indexAppleProduct:   nil,
}
