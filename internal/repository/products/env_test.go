package products

import (
	"github.com/FTChinese/subscription-api/internal/repository/shared"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/patrickmn/go-cache"
)

func newTestEnv(
	dbs db.ReadWriteMyDBs,
	cache *cache.Cache,
) Env {
	return Env{
		PaywallCommon: shared.PaywallCommon{
			DBs:   dbs,
			Cache: cache,
		},
	}
}
