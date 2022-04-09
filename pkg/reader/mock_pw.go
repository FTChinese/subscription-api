//go:build !production
// +build !production

package reader

import "github.com/FTChinese/subscription-api/pkg/price"

var MockPwPriceStdYear = PaywallPrice{
	FtcPrice: price.MockPriceStdYear,
	Offers:   price.MockStdYearOffers,
}

var MockPwPriceStdMonth = PaywallPrice{
	FtcPrice: price.MockPriceStdMonth,
	Offers:   price.MockStdMonthOffers,
}

var MockPwPricePrm = PaywallPrice{
	FtcPrice: price.MockPricePrm,
	Offers:   price.MockPrmOffers,
}
