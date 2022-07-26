package repository

import (
	"github.com/FTChinese/subscription-api/lib/sq"
	"github.com/FTChinese/subscription-api/pkg/price"
)

func (repo StripeRepo) RetrievePaywallPrices(ids []string, live bool) ([]price.StripePrice, error) {
	var list = make([]price.StripePrice, 0)
	err := repo.dbs.Read.Select(
		&list,
		price.StmtListStripePrice,
		sq.FindInSetValue(ids),
		live)

	if err != nil {
		return nil, err
	}

	return list, nil
}

// RetrievePaywallCoupons retrieves active coupons of the specified prices used on paywall.
func (repo StripeRepo) RetrievePaywallCoupons(priceIDs []string, live bool) ([]price.StripeCoupon, error) {
	var list = make([]price.StripeCoupon, 0)
	err := repo.dbs.Read.Select(
		&list,
		price.StmtActiveCouponsOfPrices,
		sq.FindInSetValue(priceIDs),
		live)

	if err != nil {
		return nil, err
	}

	return list, nil
}
