package striperepo

import (
	"github.com/FTChinese/subscription-api/internal/repository/readers"
	"github.com/FTChinese/subscription-api/internal/repository/txrepo"
	"github.com/FTChinese/subscription-api/pkg/db"
	"go.uber.org/zap"
)

// Env wraps database connection
type Env struct {
	readers.Env
	client Client
	logger *zap.Logger
}

// New creates a new instance of Env.
// `sandbox` is used to determine which table to write subscription data.
func New(dbs db.ReadWriteMyDBs, client Client, logger *zap.Logger) Env {
	return Env{
		Env:    readers.New(dbs, logger),
		client: client,
		logger: logger,
	}
}

func (env Env) beginStripeTx() (txrepo.StripeTx, error) {
	tx, err := env.DBs.Write.Beginx()

	if err != nil {
		return txrepo.StripeTx{}, err
	}

	return txrepo.NewStripeTx(tx), nil
}
