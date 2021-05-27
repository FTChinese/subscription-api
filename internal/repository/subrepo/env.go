package subrepo

import (
	"github.com/FTChinese/subscription-api/internal/repository/readers"
	"github.com/FTChinese/subscription-api/internal/repository/txrepo"
	"github.com/FTChinese/subscription-api/pkg/db"
	"go.uber.org/zap"
)

// Env wraps database connection and embed
// readers.Env to access user account data.
type Env struct {
	readers.Env
	logger *zap.Logger
}

// NewEnv creates a new instance of Env.
// `sandbox` is used to determine which table to write subscription data.
func NewEnv(dbs db.ReadWriteSplit, logger *zap.Logger) Env {
	return Env{
		Env:    readers.New(dbs, logger),
		logger: logger,
	}
}

func (env Env) BeginOrderTx() (txrepo.OrderTx, error) {
	tx, err := env.DBs.Delete.Beginx()

	if err != nil {
		return txrepo.OrderTx{}, err
	}

	return txrepo.NewOrderTx(tx), nil
}
