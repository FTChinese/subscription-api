package iaprepo

import (
	"context"
	"github.com/FTChinese/subscription-api/internal/repository/txrepo"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

var ctx = context.Background()

type Env struct {
	dbs    db.ReadWriteMyDBs
	rdb    *redis.Client
	logger *zap.Logger
}

func New(dbs db.ReadWriteMyDBs, rdb *redis.Client, logger *zap.Logger) Env {
	return Env{
		dbs:    dbs,
		rdb:    rdb,
		logger: logger,
	}
}

// beginIAPTx starts a transaction.
// NOTE: here the sandbox is different from the environment
// field send by apple. It only determines whether the
// sandbox db should be used and is determined by
// the CLI argument `-sandbox`.
// All messages from apple is save in production DBs.
func (env Env) beginIAPTx() (txrepo.IAPTx, error) {
	tx, err := env.dbs.Delete.Beginx()

	if err != nil {
		return txrepo.IAPTx{}, err
	}

	return txrepo.NewIAPTx(tx), nil
}
