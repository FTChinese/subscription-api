package db

import (
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/FTChinese/subscription-api/pkg/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var devConfig = &gorm.Config{
	Logger: logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.Info,
			Colorful:      true,
		},
	),
}

func newGormConfig(production bool) *gorm.Config {
	if production {
		return &gorm.Config{}
	}

	return devConfig
}

func NewGormDB(c config.Connect, production bool) (*gorm.DB, error) {
	dsn := buildDSN(c)
	return gorm.Open(mysql.Open(dsn), newGormConfig(production))
}

func MustNewGormDB(c config.Connect, production bool) *gorm.DB {
	db, err := NewGormDB(c, production)
	if err != nil {
		panic(err)
	}

	return db
}

type MultiGormDBs struct {
	Read   *gorm.DB
	Write  *gorm.DB
	Delete *gorm.DB
}

func MustNewMultiGormDBs(prod bool) MultiGormDBs {
	return MultiGormDBs{
		Read:   MustNewGormDB(config.MustMySQLReadConn(), prod),
		Write:  MustNewGormDB(config.MustMySQLWriteConn(), prod),
		Delete: MustNewGormDB(config.MustMySQLDeleteConn(), prod),
	}
}

func (x ReadWriteMyDBs) OpenGormDBs(prod bool) MultiGormDBs {

	return MultiGormDBs{
		Read:   mustGormOpenExistingDB(x.Read.DB, prod),
		Write:  mustGormOpenExistingDB(x.Write.DB, prod),
		Delete: mustGormOpenExistingDB(x.Delete.DB, prod),
	}
}

func mustGormOpenExistingDB(sqlDB *sql.DB, prod bool) *gorm.DB {
	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn: sqlDB,
	}), newGormConfig(prod))

	if err != nil {
		panic(err)
	}

	return db
}
