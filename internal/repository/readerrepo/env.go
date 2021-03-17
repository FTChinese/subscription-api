package readerrepo

import (
	"github.com/FTChinese/subscription-api/internal/repository/txrepo"
	"github.com/FTChinese/subscription-api/pkg/db"
	"go.uber.org/zap"
)

type Env struct {
	dbs    db.ReadWriteSplit
	logger *zap.Logger
}

func NewEnv(dbs db.ReadWriteSplit, logger *zap.Logger) Env {
	return Env{
		dbs:    dbs,
		logger: logger,
	}
}

func (env Env) BeginTx() (txrepo.MemberTx, error) {
	tx, err := env.dbs.Write.Beginx()

	if err != nil {
		return txrepo.MemberTx{}, err
	}

	return txrepo.NewMemberTx(tx), nil
}
