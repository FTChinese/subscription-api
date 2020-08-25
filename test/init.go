package test

import (
	"github.com/FTChinese/go-rest/postoffice"
	"github.com/FTChinese/subscription-api/pkg/ali"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/wechat"
	"github.com/FTChinese/subscription-api/pkg/wxlogin"
	"github.com/jmoiron/sqlx"
	"github.com/patrickmn/go-cache"
)

const (
	MyFtcID    = "e1a1f5c0-0e23-11e8-aa75-977ba2bcc6ae"
	MyFtcEmail = "neefrankie@163.com"
	MyUnionID  = "ogfvwjk6bFqv2yQpOrac0J3PqA0o"
	MyEmail    = "neefrankie@gmail.com"
)

var (
	CFG         = config.NewBuildConfig(false, false)
	DB          *sqlx.DB
	Postman     postoffice.PostOffice
	Cache       *cache.Cache
	WxOAuthApp  wxlogin.OAuthApp
	WxPayApp    wechat.PayApp = wechat.MustNewPayApp("wxapp.native_app")
	WxPayClient wechat.Client
	AliApp      ali.App
)

func init() {
	DB = db.MustNewDB(CFG.MustGetDBConn(""))
	Postman = postoffice.New(config.MustGetHanqiConn())
	Cache = cache.New(cache.DefaultExpiration, 0)
	WxOAuthApp = wxlogin.MustNewOAuthApp("wxapp.native_app")
	WxPayApp = wechat.MustNewPayApp("wxapp.native_app")
	WxPayClient = wechat.NewClient(WxPayApp)
	AliApp = ali.MustInitApp()
}
