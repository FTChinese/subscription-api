package products

import (
	"github.com/FTChinese/subscription-api/internal/repository/shared"
	"github.com/FTChinese/subscription-api/internal/repository/txrepo"
)

// Env extends PaywallCommon, mostly with db write capabilities.
type Env struct {
	shared.PaywallCommon
}

func New(base shared.PaywallCommon) Env {
	return Env{
		PaywallCommon: base,
	}
}

func (env Env) ClearCache() {
	env.Cache.Flush()
}

func (env Env) beginPriceTx() (txrepo.PriceTx, error) {
	tx, err := env.DBs.Write.Beginx()
	if err != nil {
		return txrepo.PriceTx{}, err
	}

	return txrepo.NewPriceTx(tx), nil
}

func (env Env) beginProductTx() (txrepo.ProductTx, error) {
	tx, err := env.DBs.Write.Beginx()

	if err != nil {
		return txrepo.ProductTx{}, err
	}

	return txrepo.NewProductTx(tx), nil
}
