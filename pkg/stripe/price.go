package stripe

import (
	"github.com/FTChinese/subscription-api/pkg/product"
	"github.com/guregu/null"
	"github.com/stripe/stripe-go/v72"
)

type PricePreset struct {
	product.Edition
	Live bool
}

var presetPrices = map[string]PricePreset{
	"plan_FXZYLOEbcvj5Tx": {
		Edition: product.NewStdMonthEdition(),
		Live:    true,
	},
	"plan_FXZZUEDpToPlZK": {
		Edition: product.NewStdYearEdition(),
		Live:    true,
	},
	"plan_FXZbv1cDTsUKOg": {
		Edition: product.NewPremiumEdition(),
		Live:    true,
	},
	"plan_FOdgPTznDwHU4i": {
		Edition: product.NewStdMonthEdition(),
		Live:    false,
	},
	"plan_FOdfeaqzczp6Ag": {
		Edition: product.NewStdYearEdition(),
		Live:    false,
	},
	"plan_FOde0uAr0V4WmT": {
		Edition: product.NewPremiumEdition(),
		Live:    false,
	},
}

type Price struct {
	ID string `json:"id"`
	product.Edition
	Active     bool        `json:"active"`
	Currency   string      `json:"currency"`
	LiveMode   bool        `json:"liveMode"`
	Nickname   null.String `json:"nickname"`
	ProductID  string      `json:"productId"`
	UnitAmount int64       `json:"unitAmount"`
	Created    int64       `json:"created"`
}

func NewPrice(preset PricePreset, price *stripe.Price) Price {
	return Price{
		ID:         price.ID,
		Edition:    preset.Edition,
		Active:     price.Active,
		Currency:   string(price.Currency),
		LiveMode:   price.Livemode,
		Nickname:   null.NewString(price.Nickname, price.Nickname != ""),
		ProductID:  price.Product.ID,
		UnitAmount: price.UnitAmount,
		Created:    price.Created,
	}
}
