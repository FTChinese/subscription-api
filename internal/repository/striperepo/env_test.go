package striperepo

import (
	"github.com/FTChinese/subscription-api/internal/repository/stripeclient"
	"github.com/FTChinese/subscription-api/pkg/db"
	"go.uber.org/zap"
)

// newTestEnv creates a new instance of Env.
// `sandbox` is used to determine which table to write subscription data.
func newTestEnv(
	dbs db.ReadWriteMyDBs,
	logger *zap.Logger,
) Env {
	return Env{
		dbs:    dbs,
		client: stripeclient.New(false, logger),
		logger: logger,
	}
}
