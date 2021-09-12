// +build !production

package db

import (
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/jmoiron/sqlx"
)

func MockMySQL() ReadWriteMyDBs {
	config.MustSetupViperV2(config.MustReadConfigFile())
	return MustNewMyDBs(false)
}

func MockTx() *sqlx.Tx {
	return MockMySQL().Write.MustBegin()
}
