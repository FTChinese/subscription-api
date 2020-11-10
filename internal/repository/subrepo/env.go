package subrepo

import (
	"github.com/FTChinese/subscription-api/internal/repository/txrepo"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// Env wraps database connection
type Env struct {
	db     *sqlx.DB
	logger *zap.Logger
}

// NewEnv creates a new instance of Env.
// `sandbox` is used to determine which table to write subscription data.
func NewEnv(db *sqlx.DB, logger *zap.Logger) Env {
	return Env{
		db:     db,
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
