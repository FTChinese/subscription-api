package subrepo

import (
	"github.com/jmoiron/sqlx"
	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
	"gitlab.com/ftchinese/subscription-api/models/util"
	"gitlab.com/ftchinese/subscription-api/repository/txrepo"
)

// SubEnv wraps database connection
type SubEnv struct {
	util.BuildConfig
	db    *sqlx.DB
	cache *cache.Cache
}

// NewSubEnv creates a new instance of SubEnv.
// `sandbox` is used to determine which table to write subscription data.
func NewSubEnv(db *sqlx.DB, c *cache.Cache, config util.BuildConfig) SubEnv {
	return SubEnv{
		BuildConfig: config,
		db:          db,
		cache:       c,
	}
}

func (env SubEnv) BeginOrderTx() (txrepo.OrderTx, error) {
	tx, err := env.db.Beginx()

	if err != nil {
		return txrepo.OrderTx{}, err
	}

	return txrepo.NewOrderTx(tx, env.UseSandboxDB()), nil
}

var logger = logrus.
	WithField("project", "subscription-api").
	WithField("package", "repository")

const (
	keyPromo = "promotionSchedule"
)
