package subrepo

import (
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/repository/txrepo"
	"github.com/jmoiron/sqlx"
	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
)

// SubEnv wraps database connection
type SubEnv struct {
	config.BuildConfig
	db    *sqlx.DB
	cache *cache.Cache
}

// NewSubEnv creates a new instance of SubEnv.
// `sandbox` is used to determine which table to write subscription data.
func NewSubEnv(db *sqlx.DB, c *cache.Cache, config config.BuildConfig) SubEnv {
	return SubEnv{
		BuildConfig: config,
		db:          db,
		cache:       c,
	}
}

func (env SubEnv) BeginOrderTx() (txrepo.MemberTx, error) {
	tx, err := env.db.Beginx()

	if err != nil {
		return txrepo.MemberTx{}, err
	}

	return txrepo.NewMemberTx(tx, env.BuildConfig), nil
}

var logger = logrus.
	WithField("project", "subscription-api").
	WithField("package", "repository")

const (
	keyPromo = "promotionSchedule"
)
