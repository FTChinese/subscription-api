package products

import (
	"github.com/jmoiron/sqlx"
	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
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

var logger = logrus.
	WithField("project", "subscription-api").
	WithField("package", "repository.products")

func getLogger(name string) *logrus.Entry {
	return logger.WithField("trace", name)
}

const (
	keyPaywall = "paywall"
	keyPricing = "pricing_plans"
)
