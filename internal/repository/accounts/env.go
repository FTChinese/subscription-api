package accounts

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

func (env Env) beginAccountTx() (txrepo.AccountTx, error) {
	tx, err := env.dbs.Delete.Beginx()
	if err != nil {
		return txrepo.AccountTx{}, err
	}

	return txrepo.NewAccountTx(tx), nil
}

func (env Env) beginUnlinkTx() (txrepo.UnlinkTx, error) {
	tx, err := env.dbs.Write.Beginx()
	if err != nil {
		return txrepo.UnlinkTx{}, err
	}

	return txrepo.UnlinkTx{Tx: tx}, nil
}

func (env Env) beginMemberTx() (txrepo.SharedTx, error) {
	tx, err := env.dbs.Write.Beginx()
	if err != nil {
		return txrepo.SharedTx{}, err
	}

	return txrepo.NewSharedTx(tx), nil
}
