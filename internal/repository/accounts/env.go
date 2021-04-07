package accounts

import (
	"github.com/FTChinese/subscription-api/pkg/db"
	"go.uber.org/zap"
)

type Env struct {
	dbs    db.ReadWriteSplit
	logger *zap.Logger
}

func New(dbs db.ReadWriteSplit, logger *zap.Logger) Env {
	return Env{
		dbs:    dbs,
		logger: logger,
	}
}

func (env Env) beginAccountTx() (AccountTx, error) {
	tx, err := env.dbs.Delete.Beginx()
	if err != nil {
		return AccountTx{}, err
	}

	return NewAccountTx(tx), nil
}
