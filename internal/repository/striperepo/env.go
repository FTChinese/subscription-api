package striperepo

import (
	"github.com/FTChinese/subscription-api/internal/repository/shared"
	"github.com/FTChinese/subscription-api/internal/repository/txrepo"
	"go.uber.org/zap"
)

// Env wraps database connection
type Env struct {
	shared.ReaderBaseRepo
	shared.StripeBaseRepo
	logger *zap.Logger
}

func New(readerBase shared.ReaderBaseRepo, stripeBase shared.StripeBaseRepo, logger *zap.Logger) Env {
	return Env{
		ReaderBaseRepo: readerBase,
		StripeBaseRepo: stripeBase,
		logger:         logger,
	}
}

func (env Env) beginStripeTx() (txrepo.StripeTx, error) {
	tx, err := env.DBs.Write.Beginx()

	if err != nil {
		return txrepo.StripeTx{}, err
	}

	return txrepo.NewStripeTx(tx), nil
}
