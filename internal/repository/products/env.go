package products

import (
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/patrickmn/go-cache"
)

type Env struct {
	dbs   db.ReadWriteSplit
	cache *cache.Cache
}

func NewEnv(dbs db.ReadWriteSplit, cache *cache.Cache) Env {
	return Env{
		dbs:   dbs,
		cache: cache,
	}
}

const (
	keyPaywall = "paywall"
	keyPricing = "pricing_plans"
)
