package products

import (
	"github.com/FTChinese/subscription-api/internal/repository/txrepo"
	"github.com/FTChinese/subscription-api/pkg/db"
)

// Env extends PaywallCommon, mostly with db write capabilities.
type Env struct {
	dbs db.ReadWriteMyDBs
}

func New(dbs db.ReadWriteMyDBs) Env {
	return Env{
		dbs: dbs,
	}
}

func (env Env) beginPriceTx() (txrepo.PriceTx, error) {
	tx, err := env.dbs.Write.Beginx()
	if err != nil {
		return txrepo.PriceTx{}, err
	}

	return txrepo.NewPriceTx(tx), nil
}

func (env Env) beginProductTx() (txrepo.ProductTx, error) {
	tx, err := env.dbs.Write.Beginx()

	if err != nil {
		return txrepo.ProductTx{}, err
	}

	return txrepo.NewProductTx(tx), nil
}
