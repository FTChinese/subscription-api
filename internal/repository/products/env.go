package products

import (
	"github.com/jmoiron/sqlx"
	"github.com/patrickmn/go-cache"
)

type Env struct {
	db    *sqlx.DB
	cache *cache.Cache
}

func NewEnv(db *sqlx.DB, cache *cache.Cache) Env {
	return Env{
		db:    db,
		cache: cache,
	}
}

const (
	keyPaywall = "paywall"
	keyPricing = "pricing_plans"
)
