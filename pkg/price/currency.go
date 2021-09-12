package price

import (
	"database/sql/driver"
	"errors"
)

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

func (x Currency) Symbol() string {
	return currencySymbolMap[x]
}

func (x *Currency) Scan(src interface{}) error {
	if src == nil {
		*x = ""
		return nil
	}

	switch s := src.(type) {
	case []byte:
		*x = Currency(s)
		return nil

	default:
		return errors.New("incompatible type to scan to OfferKind")
	}
}

func (x Currency) Value() (driver.Value, error) {
	if x == "" {
		return nil, nil
	}

	return string(x), nil
}
