package subrepo

import (
	"github.com/FTChinese/subscription-api/internal/repository/txrepo"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/jmoiron/sqlx"
	"github.com/patrickmn/go-cache"
)

// Env wraps database connection
type Env struct {
	config.BuildConfig
	db    *sqlx.DB
	cache *cache.Cache
}

// NewEnv creates a new instance of Env.
// `sandbox` is used to determine which table to write subscription data.
func NewEnv(db *sqlx.DB, c *cache.Cache) Env {
	return Env{
		db:    db,
		cache: c,
	}
}

func (env Env) BeginOrderTx() (txrepo.MemberTx, error) {
	tx, err := env.db.Beginx()

	if err != nil {
		return txrepo.MemberTx{}, err
	}

	return txrepo.NewMemberTx(tx), nil
}
