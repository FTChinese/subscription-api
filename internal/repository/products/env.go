package products

import (
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/patrickmn/go-cache"
)

type Env struct {
	dbs   db.ReadWriteMyDBs
	cache *cache.Cache
}

func NewEnv(dbs db.ReadWriteMyDBs, cache *cache.Cache) Env {
	return Env{
		dbs:   dbs,
		cache: cache,
	}
}

func getPaywallCacheKey(live bool) string {
	return "paywall_" + ids.GetBoolKey(live)
}
