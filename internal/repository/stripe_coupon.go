package repository

import (
	"github.com/FTChinese/subscription-api/internal/pkg/stripe"
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

// ListPriceCoupons retrieves all coupons of a price.
func (repo StripeRepo) ListPriceCoupons(priceID string, activeOnly bool) ([]price.StripeCoupon, error) {
	var stmt = price.StmtPriceAllCoupons
	if activeOnly {
		stmt = price.StmtPriceActiveCoupons
	}

	var list = make([]price.StripeCoupon, 0)
	err := repo.dbs.Read.Select(
		&list,
		stmt,
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

func (repo StripeRepo) InsertCouponRedeemed(r stripe.CouponRedeemed) error {
	_, err := repo.dbs.Write.NamedExec(
		stripe.StmtInsertCouponRedeemed,
		r)

	return err
}

func (repo StripeRepo) LatestCouponApplied(invoiceID string) (stripe.CouponRedeemed, error) {
	var r stripe.CouponRedeemed
	err := repo.dbs.Read.Get(
		&r,
		stripe.StmtLatestCouponRedeemed,
		invoiceID)

	if err != nil {
		return stripe.CouponRedeemed{}, err
	}

	return r, nil
}
