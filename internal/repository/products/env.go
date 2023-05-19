package products

import (
	"github.com/FTChinese/subscription-api/pkg/db"
)

// Env extends PaywallCommon, mostly with db write capabilities.
type Env struct {
	dbs db.ReadWriteMyDBs
}

func New(dbs db.ReadWriteMyDBs) Env {
	return Env{
		dbs: dbs,
	}
}
