package striperepo

import (
	"github.com/FTChinese/subscription-api/internal/repository/txrepo"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// Env wraps database connection
type Env struct {
	db     *sqlx.DB
	dbs    db.ReadWriteSplit
	client Client
	logger *zap.Logger
}

// NewSubEnv creates a new instance of SubEnv.
// `sandbox` is used to determine which table to write subscription data.
func NewEnv(dbs db.ReadWriteSplit, client Client, logger *zap.Logger) Env {
	return Env{
		dbs:    dbs,
		client: client,
		logger: logger,
	}
}

func (env Env) beginAccountTx() (txrepo.AccountTx, error) {
	tx, err := env.dbs.Write.Beginx()

	if err != nil {
		return txrepo.AccountTx{}, err
	}

	return txrepo.NewAccountTx(tx), nil
}

func (env Env) beginSubsTx() (txrepo.MemberTx, error) {
	tx, err := env.dbs.Write.Beginx()

	if err != nil {
		return txrepo.MemberTx{}, err
	}

	return txrepo.NewMemberTx(tx), nil
}
