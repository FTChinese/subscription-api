//go:build !production
// +build !production

package db

import (
	"github.com/FTChinese/subscription-api/faker"
	"github.com/jmoiron/sqlx"
)

func MockMySQL() ReadWriteMyDBs {
	faker.MustSetupViper()
	return MustNewMyDBs(false)
}

func MockTx() *sqlx.Tx {
	return MockMySQL().Write.MustBegin()
}
