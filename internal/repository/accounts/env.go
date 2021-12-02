package accounts

import (
	"github.com/FTChinese/subscription-api/internal/repository/shared"
	"github.com/FTChinese/subscription-api/internal/repository/txrepo"
	"go.uber.org/zap"
)

type Env struct {
	shared.ReaderBaseRepo
	logger *zap.Logger
}

func New(baseRepo shared.ReaderBaseRepo, logger *zap.Logger) Env {
	return Env{
		ReaderBaseRepo: baseRepo,
		logger:         logger,
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

func (env Env) beginMemberTx() (txrepo.SharedTx, error) {
	tx, err := env.DBs.Write.Beginx()
	if err != nil {
		return txrepo.SharedTx{}, err
	}

	return txrepo.NewSharedTx(tx), nil
}
