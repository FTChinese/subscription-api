// +build !production

package test

import (
	"github.com/FTChinese/subscription-api/pkg/ali"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/wechat"
	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
	"github.com/patrickmn/go-cache"
)

const (
	MyEmail = "neefrankie@gmail.com"
)

var (
	CFG      = config.NewBuildConfig(false, false)
	DB       *sqlx.DB
	SplitDB  db.ReadWriteSplit
	Redis    *redis.Client
	Cache    *cache.Cache
	WxPayApp wechat.PayApp
	AliApp   ali.App
)

func init() {
	config.MustSetupViper()

	DB = db.MustNewMySQL(CFG.MustGetDBConn(""))
	SplitDB = db.ReadWriteSplit{
		Read:   DB,
		Write:  DB,
		Delete: DB,
	}
	Redis = db.NewRedis(config.MustRedisAddress().Pick(false))
	Cache = cache.New(cache.DefaultExpiration, 0)
	WxPayApp = wechat.MustNewPayApp("wxapp.native_app")
	AliApp = ali.MustInitApp()
}
