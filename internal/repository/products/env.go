package products

import (
	"github.com/FTChinese/subscription-api/internal/repository/txrepo"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/patrickmn/go-cache"
)

type Env struct {
	dbs   db.ReadWriteMyDBs
	cache *cache.Cache
}

func NewEnv(
	dbs db.ReadWriteMyDBs,
	cache *cache.Cache,
) Env {
	return Env{
		dbs:   dbs,
		cache: cache,
	}
}

func getPaywallCacheKey(live bool) string {
	return "paywall_" + ids.GetBoolKey(live)
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
