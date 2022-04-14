package repository

import (
	"github.com/FTChinese/subscription-api/lib/sq"
	"github.com/FTChinese/subscription-api/pkg/price"
)

func (repo StripeRepo) RetrievePaywallPrices(ids []string) ([]price.StripePrice, error) {
	var list = make([]price.StripePrice, 0)
	err := repo.dbs.Read.Select(
		&list,
		price.StmtListStripePrice,
		sq.FindInSetValue(ids))

	if err != nil {
		return nil, err
	}

	return list, nil
}

func (repo StripeRepo) RetrievePaywallCoupons(priceIDs []string) ([]price.StripeCoupon, error) {
	var list = make([]price.StripeCoupon, 0)
	err := repo.dbs.Read.Select(
		&list,
		price.StmtCouponsOfPrices,
		sq.FindInSetValue(priceIDs))

	if err != nil {
		return nil, err
	}

	return list, nil
}
