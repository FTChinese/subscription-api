package repository

import (
	"github.com/FTChinese/subscription-api/pkg/db"
	"go.uber.org/zap"
)

type StripeRepo struct {
	dbs    db.ReadWriteMyDBs
	Logger *zap.Logger
}

func NewStripeRepo(dbs db.ReadWriteMyDBs, logger *zap.Logger) StripeRepo {
	return StripeRepo{
		dbs:    dbs,
		Logger: logger,
	}
}

func (repo StripeRepo) BeginStripeTx() (StripeTx, error) {
	tx, err := repo.dbs.Write.Beginx()

	if err != nil {
		return StripeTx{}, err
	}

	return NewStripeTx(tx), nil
}
