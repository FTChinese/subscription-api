package subrepo

import (
	"github.com/jmoiron/sqlx"
	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
	"gitlab.com/ftchinese/subscription-api/models/util"
)

// SubEnv wraps database connection
type SubEnv struct {
	util.BuildConfig
	db    *sqlx.DB
	cache *cache.Cache
}

// New creates a new instance of SubEnv.
// `sandbox` is used to determine which table to write subscription data.
func New(db *sqlx.DB, c *cache.Cache, b util.BuildConfig) SubEnv {
	return SubEnv{
		BuildConfig: b,
		db:          db,
		cache:       c,
	}
}

func (env SubEnv) BeginOrderTx() (OrderTx, error) {
	tx, err := env.db.Beginx()

	if err != nil {
		return OrderTx{}, err
	}

	return OrderTx{
		tx:      tx,
		sandbox: env.Sandbox,
	}, nil
}

var logger = logrus.
	WithField("project", "subscription-api").
	WithField("package", "repository")

const (
	keyPromo = "promotionSchedule"
)
