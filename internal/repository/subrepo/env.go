package subrepo

import (
	"github.com/FTChinese/subscription-api/internal/repository/shared"
	"github.com/FTChinese/subscription-api/internal/repository/txrepo"
	"github.com/FTChinese/subscription-api/pkg/db"
	"go.uber.org/zap"
)

// Env wraps database connection and embed
// readers.Env to access user account data.
type Env struct {
	shared.ReaderBaseRepo
	logger *zap.Logger
}

// NewEnv creates a new instance of Env.
// `sandbox` is used to determine which table to write subscription data.
// Deprecated
func NewEnv(dbs db.ReadWriteMyDBs, logger *zap.Logger) Env {
	return Env{
		ReaderBaseRepo: shared.NewReaderBaseRepo(dbs),
		logger:         logger,
	}
}

func New(baseRepo shared.ReaderBaseRepo, logger *zap.Logger) Env {
	return Env{
		ReaderBaseRepo: baseRepo,
		logger:         logger,
	}
}

func (env Env) BeginOrderTx() (txrepo.OrderTx, error) {
	tx, err := env.DBs.Delete.Beginx()

	if err != nil {
		return txrepo.OrderTx{}, err
	}

	return txrepo.NewOrderTx(tx), nil
}
