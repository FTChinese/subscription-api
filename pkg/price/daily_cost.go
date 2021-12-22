package price

import "github.com/FTChinese/subscription-api/pkg/conv"

type DailyCost struct {
	Holder   string
	Replacer string
}

func NewDailyCostOfYear(price float64) DailyCost {
	return DailyCost{
		Holder:   "{{dailyAverageOfYear}}",
		Replacer: conv.FormatMoney(price / 365),
	}
}

func NewDailyCostOfMonth(price float64) DailyCost {
	return DailyCost{
		Holder:   "{{dailyAverageOfMonth}}",
		Replacer: conv.FormatMoney(price / 30),
	}
}
