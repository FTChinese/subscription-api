package products

import (
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/reader"
)

func (env Env) CreateDiscount(d price.Discount) error {

	_, err := env.dbs.Write.NamedExec(
		price.StmtCreateDiscount,
		d)
	if err != nil {
		return err
	}

	return nil
}

// UpdateDiscount to change its status to paused or cancelled.
func (env Env) UpdateDiscount(d price.Discount) error {
	_, err := env.dbs.Write.NamedExec(price.StmtUpdateDiscount, d)
	if err != nil {
		return err
	}

	return nil
}

// ListActiveDiscounts retrieves a list of active discounts under a price.
// Used when we want to save all valid discounts in the price's row
// so that we won't bother to do complicated retrieval when
// user is requesting a price.
func (env Env) ListActiveDiscounts(priceID string, live bool) ([]price.Discount, error) {
	var list = make([]price.Discount, 0)
	// Use write db to circumvent intermittent db sync failures.
	err := env.dbs.Write.Select(
		&list,
		price.StmtListPriceActiveDiscounts,
		priceID,
		live)
	if err != nil {
		return nil, err
	}

	return list, nil
}

// ListDiscounts retrieves all discounts under a price.
// It seems this is not of much use.
func (env Env) ListDiscounts(priceID string, live bool) ([]price.Discount, error) {
	var list = make([]price.Discount, 0)
	err := env.dbs.Read.Select(
		&list,
		price.StmtListDiscountsOfPrice,
		priceID,
		live)
	if err != nil {
		return nil, err
	}

	return list, nil
}

// ArchivePriceDiscounts turns all discount under a price into
// cancelled mode.
func (env Env) ArchivePriceDiscounts(p reader.PaywallPrice) error {
	_, err := env.dbs.Write.NamedExec(
		price.StmtArchivePriceDiscounts,
		p)
	if err != nil {
		return err
	}

	return nil
}
