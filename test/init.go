package test

import (
	"database/sql"
	"github.com/FTChinese/go-rest/postoffice"
	"github.com/patrickmn/go-cache"
	"github.com/spf13/viper"
	"gitlab.com/ftchinese/subscription-api/paywall"
	"gitlab.com/ftchinese/subscription-api/util"
	"gitlab.com/ftchinese/subscription-api/wechat"
	"gitlab.com/ftchinese/subscription-api/wxlogin"
)

const (
	MyFtcID    = "e1a1f5c0-0e23-11e8-aa75-977ba2bcc6ae"
	MyFtcEmail = "neefrankie@163.com"
	MyUnionID  = "ogfvwjk6bFqv2yQpOrac0J3PqA0o"
	MyEmail    = "neefrankie@gmail.com"
)

var DB *sql.DB
var Postman postoffice.Postman
var Cache *cache.Cache
var WxOAuthApp wxlogin.WxApp
var WxPayApp wechat.PayApp
var WxPayClient wechat.Client
var StripeKey string

var YearlyStandard = paywall.GetFtcPlans(false)["standard_year"]
var YearlyPremium = paywall.GetFtcPlans(false)["premium_year"]

func init() {
	viper.SetConfigName("api")
	viper.AddConfigPath("$HOME/config")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	//DB, err = sql.Open("mysql", "sampadm:secret@unix(/tmp/mysql.sock)/")
	//if err != nil {
	//	panic(err)
	//}
	var dbConn util.Conn
	err = viper.UnmarshalKey("mysql.dev", &dbConn)
	if err != nil {
		panic(err)
	}

	DB, err = util.NewDB(dbConn)
	if err != nil {
		panic(err)
	}

	var conn util.Conn
	err = viper.UnmarshalKey("hanqi", &conn)
	if err != nil {
		panic(err)
	}
	Postman = postoffice.New(conn.Host, conn.Port, conn.User, conn.Pass)

	Cache = cache.New(cache.DefaultExpiration, 0)

	err = viper.UnmarshalKey("wxapp.m_subs", &WxOAuthApp)
	if err != nil {
		panic(err)
	}

	err = viper.UnmarshalKey("wxapp.m_subs", &WxPayApp)
	if err != nil {
		panic(err)
	}

	WxPayClient = wechat.NewClient(WxPayApp)

	StripeKey = viper.GetString("stripe.test_secret_key")
}
