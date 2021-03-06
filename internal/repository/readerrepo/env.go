package readerrepo

import (
	"github.com/FTChinese/subscription-api/internal/repository/txrepo"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type Env struct {
	//config.BuildConfig
	db     *sqlx.DB
	logger *zap.Logger
}

func NewEnv(db *sqlx.DB, logger *zap.Logger) Env {
	return Env{
		db:     db,
		logger: logger,
	}
}

func (env Env) BeginTx() (txrepo.MemberTx, error) {
	tx, err := env.db.Beginx()

	if err != nil {
		return txrepo.MemberTx{}, err
	}

	return txrepo.NewMemberTx(tx), nil
}
