//go:build !production
// +build !production

package pw

import "github.com/FTChinese/subscription-api/pkg/price"

var MockPwPriceStdYear = PaywallPrice{
	Price:  price.MockPriceStdYear,
	Offers: price.MockStdYearOffers,
}

var MockPwPriceStdMonth = PaywallPrice{
	Price:  price.MockPriceStdMonth,
	Offers: price.MockStdMonthOffers,
}

var MockPwPricePrm = PaywallPrice{
	Price:  price.MockPricePrm,
	Offers: price.MockPrmOffers,
}
