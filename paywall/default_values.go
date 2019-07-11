package paywall

import "github.com/FTChinese/go-rest/enum"

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
	},
	"standard_month": Plan{
		Tier:        enum.TierStandard,
		Cycle:       enum.CycleMonth,
		ListPrice:   28.00,
		NetPrice:    28.00,
		Description: "FT中文网 - 月度标准会员",
	},
	"premium_year": Plan{
		Tier:        enum.TierPremium,
		Cycle:       enum.CycleYear,
		ListPrice:   1998.00,
		NetPrice:    1998.00,
		Description: "FT中文网 - 高端会员",
	},
}

// GetDefaultPricing returns the default pricing plans.
func GetDefaultPricing() FtcPlans {
	return defaultPlans
}

func GetDefaultBanner() Banner {
	return defaultBanner
}
