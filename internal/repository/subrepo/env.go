package subrepo

import (
	"github.com/FTChinese/subscription-api/internal/repository/txrepo"
	"github.com/FTChinese/subscription-api/pkg/db"
	"go.uber.org/zap"
)

// Env wraps database connection
type Env struct {
	dbs    db.ReadWriteSplit
	logger *zap.Logger
}

// NewEnv creates a new instance of Env.
// `sandbox` is used to determine which table to write subscription data.
func NewEnv(dbs db.ReadWriteSplit, logger *zap.Logger) Env {
	return Env{
		dbs:    dbs,
		logger: logger,
	}
}

func (env Env) BeginOrderTx() (txrepo.MemberTx, error) {
	tx, err := env.dbs.Delete.Beginx()

	if err != nil {
		return txrepo.MemberTx{}, err
	}

	return txrepo.NewMemberTx(tx), nil
}
