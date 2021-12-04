package products

import (
	"github.com/FTChinese/subscription-api/pkg/db"
)

func newTestEnv(
	dbs db.ReadWriteMyDBs,
) Env {
	return Env{
		dbs: dbs,
	}
}
