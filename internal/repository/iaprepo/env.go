package iaprepo

import (
	"github.com/FTChinese/subscription-api/internal/repository/txrepo"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type Env struct {
	db     *sqlx.DB
	logger *zap.Logger
}

func NewEnv(db *sqlx.DB, logger *zap.Logger) Env {
	return Env{
		db:     db,
		logger: logger,
	}
}

// BeginTx starts a transaction.
// NOTE: here the sandbox is different from the environment
// field send by apple. It only determines whether the
// sandbox db should be used and is determined by
// the CLI argument `-sandbox`.
// All messages from apple is save in production DBs.
func (env Env) BeginTx() (txrepo.MemberTx, error) {
	tx, err := env.db.Beginx()

	if err != nil {
		return txrepo.MemberTx{}, err
	}

	return txrepo.NewMemberTx(tx), nil
}
