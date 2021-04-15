package addons

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

func (env Env) beginAddOnTx() (txrepo.AddOnTx, error) {
	tx, err := env.dbs.Write.Beginx()

	if err != nil {
		return txrepo.AddOnTx{}, err
	}

	return txrepo.NewAddOnTx(tx), nil
}
