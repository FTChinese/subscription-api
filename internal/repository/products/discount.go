package products

import "github.com/FTChinese/subscription-api/pkg/price"

// LoadDiscount retrieve a single row of discount
func (env Env) LoadDiscount(id string) (price.Discount, error) {
	var d price.Discount
	err := env.dbs.Read.Get(
		&d,
		price.StmtSelectDiscount,
		id)
	if err != nil {
		return price.Discount{}, err
	}

	return d, nil
}

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
	err := env.dbs.Read.Select(
		&list,
		price.StmtListPriceActiveDiscounts,
		priceID,
		live)
	if err != nil {
		return nil, err
	}

	return list, nil
}

func (env Env) ListDiscounts(priceID string) ([]price.Discount, error) {
	var list = make([]price.Discount, 0)
	err := env.dbs.Read.Select(
		&list,
		price.StmtListDiscountsOfPrice,
		priceID)
	if err != nil {
		return nil, err
	}

	return list, nil
}
