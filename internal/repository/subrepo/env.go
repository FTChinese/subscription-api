package subrepo

import (
	"github.com/FTChinese/subscription-api/internal/repository/txrepo"
	"github.com/jmoiron/sqlx"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
)

// Env wraps database connection
type Env struct {
	db     *sqlx.DB
	cache  *cache.Cache
	logger *zap.Logger
}

// NewEnv creates a new instance of Env.
// `sandbox` is used to determine which table to write subscription data.
func NewEnv(db *sqlx.DB, c *cache.Cache, logger *zap.Logger) Env {
	return Env{
		db:     db,
		cache:  c,
		logger: logger,
	}
}

func (env Env) BeginOrderTx() (txrepo.MemberTx, error) {
	tx, err := env.db.Beginx()

	if err != nil {
		return txrepo.MemberTx{}, err
	}

	return txrepo.NewMemberTx(tx), nil
}
