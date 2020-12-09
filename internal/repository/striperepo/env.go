package striperepo

import (
	"github.com/FTChinese/subscription-api/internal/repository/txrepo"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// Env wraps database connection
type Env struct {
	db     *sqlx.DB
	client Client
	logger *zap.Logger
}

// NewSubEnv creates a new instance of SubEnv.
// `sandbox` is used to determine which table to write subscription data.
func NewEnv(db *sqlx.DB, client Client, logger *zap.Logger) Env {
	return Env{
		db:     db,
		client: client,
		logger: logger,
	}
}

func (env Env) beginAccountTx() (txrepo.AccountTx, error) {
	tx, err := env.db.Beginx()

	if err != nil {
		return txrepo.AccountTx{}, err
	}

	return txrepo.NewAccountTx(tx), nil
}

func (env Env) beginSubsTx() (txrepo.MemberTx, error) {
	tx, err := env.db.Beginx()

	if err != nil {
		return txrepo.MemberTx{}, err
	}

	return txrepo.NewMemberTx(tx), nil
}
