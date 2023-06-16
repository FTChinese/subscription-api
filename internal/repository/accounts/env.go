package accounts

import (
	"github.com/FTChinese/subscription-api/internal/repository/txrepo"
	"github.com/FTChinese/subscription-api/pkg/db"
	"go.uber.org/zap"
)

type Env struct {
	dbs    db.ReadWriteMyDBs // Deprecated
	logger *zap.Logger
}

// Deprecated
func New(myDBs db.ReadWriteMyDBs, logger *zap.Logger) Env {
	return Env{
		dbs:    myDBs,
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
