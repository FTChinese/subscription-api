package accounts

import (
	"github.com/FTChinese/subscription-api/internal/repository/shared"
	"github.com/FTChinese/subscription-api/pkg/db"
	"go.uber.org/zap"
)

func newTestEnv(dbs db.ReadWriteMyDBs, logger *zap.Logger) Env {
	return Env{
		ReaderBaseRepo: shared.NewReaderBaseRepo(dbs),
		logger:         logger,
	}
}
