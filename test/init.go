//go:build !production
// +build !production

package test

import (
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/ali"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/postman"
	"github.com/FTChinese/subscription-api/pkg/wechat"
	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
	"github.com/objcoding/wxpay"
	"github.com/patrickmn/go-cache"
)

var (
	DB          *sqlx.DB
	SplitDB     db.ReadWriteMyDBs
	Postman     postman.Postman
	Redis       *redis.Client
	Cache       *cache.Cache
	WxPayApp    wechat.PayApp
	WxPayClient *wxpay.Client
	AliApp      ali.App
)

func init() {
	faker.MustSetupViper()

	SplitDB = db.MustNewMyDBs(false)
	DB = SplitDB.Write
	Postman = postman.New(config.MustGetHanqiConn())
	Redis = db.NewRedis(config.MustRedisAddress().Pick(false))
	Cache = cache.New(cache.DefaultExpiration, 0)
	WxPayApp = wechat.MustNewPayApp("wxapp.app_pay")
	WxPayClient = wxpay.NewClient(wxpay.NewAccount(WxPayApp.AppID, WxPayApp.MchID, WxPayApp.APIKey, false))
	AliApp = ali.MustInitApp()
}
