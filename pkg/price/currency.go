package price

type Currency string

const (
	CurrencyCNY Currency = "cny" // Chinese Renminbi Yuan
	CurrencyEUR Currency = "eur" // Euro
	CurrencyGBP Currency = "gbp" // British Pound
	CurrencyHKD Currency = "hkd" // Hong Kong Dollar
	CurrencyJPY Currency = "jpy" // Japanese Yen
	CurrencyUSD Currency = "usd" // United States Dollar
)

var currencySymbolMap = map[Currency]string{
	CurrencyCNY: "¥",
	CurrencyEUR: "€",
	CurrencyGBP: "£",
	CurrencyHKD: "HK$",
	CurrencyJPY: "¥",
	CurrencyUSD: "US$",
}

func (c Currency) Symbol() string {
	return currencySymbolMap[c]
}
