package accounts

import (
	"github.com/FTChinese/subscription-api/internal/repository/readers"
	"github.com/FTChinese/subscription-api/internal/repository/txrepo"
	"github.com/FTChinese/subscription-api/pkg/db"
	"go.uber.org/zap"
)

type Env struct {
	readers.Env
}

func New(dbs db.ReadWriteSplit, logger *zap.Logger) Env {
	return Env{
		Env: readers.New(dbs, logger),
	}
}

func (env Env) beginAccountTx() (txrepo.AccountTx, error) {
	tx, err := env.DBs.Delete.Beginx()
	if err != nil {
		return txrepo.AccountTx{}, err
	}

	return txrepo.NewAccountTx(tx), nil
}

func (env Env) beginUnlinkTx() (txrepo.UnlinkTx, error) {
	tx, err := env.DBs.Write.Beginx()
	if err != nil {
		return txrepo.UnlinkTx{}, err
	}

	return txrepo.UnlinkTx{Tx: tx}, nil
}
