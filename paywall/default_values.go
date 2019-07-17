package paywall

import (
	"fmt"
	"github.com/FTChinese/go-rest/enum"
)

// The default banner message used on web version of pay wall.
var defaultBanner = Banner{
	Heading:    "FT中文网会员订阅服务",
	CoverURL:   "http://www.ftacademy.cn/subscription.jpg",
	SubHeading: "欢迎您",
	Content: []string{
		"希望全球视野的FT中文网，能够带您站在高海拔的地方俯瞰世界，引发您的思考，从不同的角度看到不一样的事物，见他人之未见！",
	},
}

// DefaultPlans is the default subscription. No discount.
var defaultPlans = FtcPlans{
	"standard_year": Plan{
		Tier:        enum.TierStandard,
		Cycle:       enum.CycleYear,
		ListPrice:   258.00,
		NetPrice:    258.00,
		Description: "FT中文网 - 年度标准会员",
		CycleCount:  1,
		Currency:    "￥",
		ExtraDays:   1,
	},
	"standard_month": Plan{
		Tier:        enum.TierStandard,
		Cycle:       enum.CycleMonth,
		ListPrice:   28.00,
		NetPrice:    28.00,
		Description: "FT中文网 - 月度标准会员",
		CycleCount:  1,
		Currency:    "￥",
		ExtraDays:   1,
	},
	"premium_year": Plan{
		Tier:        enum.TierPremium,
		Cycle:       enum.CycleYear,
		ListPrice:   1998.00,
		NetPrice:    1998.00,
		Description: "FT中文网 - 高端会员",
		CycleCount:  1,
		Currency:    "￥",
		ExtraDays:   1,
	},
}

var sandboxPlans = FtcPlans{
	"standard_year": Plan{
		Tier:        enum.TierStandard,
		Cycle:       enum.CycleYear,
		ListPrice:   258.00,
		NetPrice:    0.01,
		Description: "FT中文网 - 年度标准会员",
		CycleCount:  1,
		Currency:    "￥",
		ExtraDays:   1,
	},
	"standard_month": Plan{
		Tier:        enum.TierStandard,
		Cycle:       enum.CycleMonth,
		ListPrice:   28.00,
		NetPrice:    0.01,
		Description: "FT中文网 - 月度标准会员",
		CycleCount:  1,
		Currency:    "￥",
		ExtraDays:   1,
	},
	"premium_year": Plan{
		Tier:        enum.TierPremium,
		Cycle:       enum.CycleYear,
		ListPrice:   1998.00,
		NetPrice:    0.01,
		Description: "FT中文网 - 高端会员",
		CycleCount:  1,
		Currency:    "￥",
		ExtraDays:   1,
	},
}

var stripeTestPlanIDs = map[string]string{
	"standard_year":  "plan_FOdfeaqzczp6Ag",
	"standard_month": "plan_FOdgPTznDwHU4i",
	"premium_year":   "plan_FOde0uAr0V4WmT",
}

var stripeLivePlanIDs = map[string]string{}

// GetDefaultPlans returns the default pricing plans.
func GetDefaultPlans() FtcPlans {
	return defaultPlans
}

// GetSandboxPlans returns the plans used under sandbox.
func GetSandboxPlans() FtcPlans {
	return sandboxPlans
}

func GetStripePlanID(key string, live bool) (string, error) {
	var id string
	var ok bool

	if live {
		id, ok = stripeLivePlanIDs[key]
	} else {
		id, ok = stripeTestPlanIDs[key]
	}

	if !ok {
		return id, fmt.Errorf("plan for %s not found", key)
	}

	return id, nil
}

func GetDefaultBanner() Banner {
	return defaultBanner
}
