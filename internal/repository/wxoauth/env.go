package wxoauth

import (
	"github.com/FTChinese/subscription-api/pkg/db"
)

type Env struct {
	dbs db.ReadWriteSplit
}

func NewEnv(dbs db.ReadWriteSplit) Env {
	return Env{
		dbs: dbs,
	}
}
