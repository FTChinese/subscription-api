package repository

import (
	"github.com/FTChinese/subscription-api/pkg/price"
)

// ListPaywallPrices retrieves all stripe prices present on paywall.
func (repo StripeRepo) ListPaywallPrices(live bool) ([]price.StripePrice, error) {
	var prices = make([]price.StripePrice, 0)

	err := repo.dbs.Read.Select(
		&prices,
		price.StmtStripeActivePrices,
		live,
	)

	if err != nil {
		return nil, err
	}

	return prices, nil
}

// ListPaywallCoupons retrieves active coupons of the all stripe prices present on paywall.
func (repo StripeRepo) ListPaywallCoupons(live bool) ([]price.StripeCoupon, error) {
	var list = make([]price.StripeCoupon, 0)
	err := repo.dbs.Read.Select(
		&list,
		price.StmtPaywallStripeCoupons,
		live)

	if err != nil {
		return nil, err
	}

	return list, nil
}
