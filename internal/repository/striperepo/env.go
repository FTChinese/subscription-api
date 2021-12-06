package striperepo

import (
	"github.com/FTChinese/subscription-api/internal/repository/stripeclient"
	"github.com/FTChinese/subscription-api/internal/repository/txrepo"
	"github.com/FTChinese/subscription-api/pkg/db"
	"go.uber.org/zap"
)

// Env wraps database connection
type Env struct {
	dbs    db.ReadWriteMyDBs
	client stripeclient.Client
	logger *zap.Logger
}

func New(dbs db.ReadWriteMyDBs, client stripeclient.Client, logger *zap.Logger) Env {
	return Env{
		dbs:    dbs,
		client: client,
		logger: logger,
	}
}

func (env Env) beginStripeTx() (txrepo.StripeTx, error) {
	tx, err := env.dbs.Write.Beginx()

	if err != nil {
		return txrepo.StripeTx{}, err
	}

	return txrepo.NewStripeTx(tx), nil
}
