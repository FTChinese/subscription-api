package cmsrepo

import (
	"github.com/FTChinese/subscription-api/internal/repository/txrepo"
	"github.com/FTChinese/subscription-api/pkg/db"
	"go.uber.org/zap"
)

type Env struct {
	dbs    db.ReadWriteMyDBs
	logger *zap.Logger
}

func New(dbs db.ReadWriteMyDBs, logger *zap.Logger) Env {
	return Env{
		dbs:    dbs,
		logger: logger,
	}
}

func (env Env) beginMemberTx() (txrepo.SharedTx, error) {
	tx, err := env.dbs.Write.Beginx()
	if err != nil {
		return txrepo.SharedTx{}, err
	}

	return txrepo.NewSharedTx(tx), nil
}
