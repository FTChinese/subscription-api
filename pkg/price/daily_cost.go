package price

type DailyCost struct {
	Holder   string
	Replacer string
}

func NewDailyCostOfYear(price float64) DailyCost {
	return DailyCost{
		Holder:   "{{dailyAverageOfYear}}",
		Replacer: FormatMoney(price / 365),
	}
}

func NewDailyCostOfMonth(price float64) DailyCost {
	return DailyCost{
		Holder:   "{{dailyAverageOfMonth}}",
		Replacer: FormatMoney(price / 30),
	}
}
