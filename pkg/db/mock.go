//go:build !production

package db

import (
	"github.com/FTChinese/subscription-api/faker"
	"github.com/jmoiron/sqlx"
)

func MockMySQL() ReadWriteMyDBs {
	faker.MustSetupViper()
	return MustNewMyDBs()
}

func MockTx() *sqlx.Tx {
	return MockMySQL().Write.MustBegin()
}

func MockGorm() MultiGormDBs {
	faker.MustSetupViper()
	return MustNewMultiGormDBs(false)
}
