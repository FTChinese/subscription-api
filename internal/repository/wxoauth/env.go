package wxoauth

import (
	"github.com/FTChinese/subscription-api/pkg/db"
)

type Env struct {
	dbs db.ReadWriteMyDBs
}

func NewEnv(dbs db.ReadWriteMyDBs) Env {
	return Env{
		dbs: dbs,
	}
}
