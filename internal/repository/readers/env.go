package readers

import (
	"github.com/FTChinese/subscription-api/pkg/db"
	"go.uber.org/zap"
)

// Env contains shared functionalities of a reader.
// It contains methods to retrieve user's
// basic account data using various id fields.
// It also contains methods to retrieve membership
// using various ids.
type Env struct {
	DBs    db.ReadWriteSplit
	Logger *zap.Logger
}

func New(dbs db.ReadWriteSplit, logger *zap.Logger) Env {
	return Env{
		DBs:    dbs,
		Logger: logger,
	}
}
