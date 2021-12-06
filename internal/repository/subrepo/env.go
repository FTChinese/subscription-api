package subrepo

import (
	"github.com/FTChinese/subscription-api/internal/repository/txrepo"
	"github.com/FTChinese/subscription-api/pkg/db"
	"go.uber.org/zap"
)

// Env wraps database connection and embed
// readers.Env to access user account data.
type Env struct {
	dbs    db.ReadWriteMyDBs
	logger *zap.Logger
}

// New creates a new instance of Env.
func New(dbs db.ReadWriteMyDBs, logger *zap.Logger) Env {
	return Env{
		dbs:    dbs,
		logger: logger,
	}
}

func (env Env) BeginOrderTx() (txrepo.OrderTx, error) {
	tx, err := env.dbs.Delete.Beginx()

	if err != nil {
		return txrepo.OrderTx{}, err
	}

	return txrepo.NewOrderTx(tx), nil
}
