package repository

import (
	"github.com/FTChinese/subscription-api/pkg/price"
)

func (repo StripeRepo) UpsertCoupon(c price.StripeCoupon) error {
	_, err := repo.dbs.Write.NamedExec(
		price.StmtUpsertCoupon,
		c)

	return err
}

func (repo StripeRepo) RetrieveCoupon(id string) (price.StripeCoupon, error) {
	var c price.StripeCoupon
	err := repo.dbs.Read.Get(
		&c,
		price.StmtRetrieveCoupon,
		id)

	if err != nil {
		return price.StripeCoupon{}, err
	}

	return c, nil
}

func (repo StripeRepo) RetrieveActiveCouponsOfPrice(priceID string) ([]price.StripeCoupon, error) {
	var list = make([]price.StripeCoupon, 0)
	err := repo.dbs.Read.Select(
		&list,
		price.StmtActiveCouponsOfPrice,
		priceID)

	if err != nil {
		return nil, err
	}

	return list, nil
}

func (repo StripeRepo) UpdateCouponStatus(c price.StripeCoupon) error {
	_, err := repo.dbs.Write.NamedExec(
		price.StmtChangeCouponStatus,
		c)

	return err
}
