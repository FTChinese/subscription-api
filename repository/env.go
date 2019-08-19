package repository

import (
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"gitlab.com/ftchinese/subscription-api/query"

	"github.com/patrickmn/go-cache"
)

type BuildConfig struct {
	Sandbox    bool
	Production bool
}

func (c BuildConfig) Live() bool {
	return c.Production && !c.Sandbox
}

// Env wraps database connection
type Env struct {
	BuildConfig
	db    *sqlx.DB
	cache *cache.Cache
	query query.Builder
}

// New creates a new instance of Env.
// `sandbox` is used to determine which table to write subscription data.
func New(db *sqlx.DB, c *cache.Cache, b BuildConfig) Env {
	return Env{
		BuildConfig: b,
		db:          db,
		cache:       c,
		query:       query.NewBuilder(b.Sandbox),
	}
}

func (env Env) BeginOrderTx() (OrderTx, error) {
	tx, err := env.db.Begin()

	if err != nil {
		return OrderTx{}, err
	}

	return OrderTx{
		tx:    tx,
		live:  env.Live(),
		query: env.query,
	}, nil
}

var logger = logrus.
	WithField("project", "subscription-api").
	WithField("package", "repository")

const (
	keyPromo = "promotionSchedule"
)
