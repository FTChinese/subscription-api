package test

import (
	"github.com/FTChinese/go-rest/postoffice"
	"github.com/jmoiron/sqlx"
	"github.com/patrickmn/go-cache"
	"github.com/spf13/viper"
	"gitlab.com/ftchinese/subscription-api/models/plan"
	"gitlab.com/ftchinese/subscription-api/models/util"
	"gitlab.com/ftchinese/subscription-api/models/wechat"
	"gitlab.com/ftchinese/subscription-api/models/wxlogin"
)

const (
	MyFtcID    = "e1a1f5c0-0e23-11e8-aa75-977ba2bcc6ae"
	MyFtcEmail = "neefrankie@163.com"
	MyUnionID  = "ogfvwjk6bFqv2yQpOrac0J3PqA0o"
	MyEmail    = "neefrankie@gmail.com"
)

var YearlyStandard = plan.GetFtcPlans()["standard_year"]
var YearlyPremium = plan.GetFtcPlans()["premium_year"]
var YearlyStandardLive = plan.GetFtcPlans()["standard_year"]
var YearlyPremiumLive = plan.GetFtcPlans()["premium_year"]

var (
	DB          *sqlx.DB
	Postman     postoffice.Postman
	Cache       *cache.Cache
	WxOAuthApp  wxlogin.WxApp
	WxPayApp    wechat.PayApp
	WxPayClient wechat.Client
	StripeKey   string

	PlanStandardMonthly, _ = plan.FindFtcPlan("standard_month")
	PlanStandardYearly, _  = plan.FindFtcPlan("standard_year")
	PlanPremiumYearly, _   = plan.FindFtcPlan("premium_year")
)

func init() {
	viper.SetConfigName("api")
	viper.AddConfigPath("$HOME/config")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

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
	err = viper.UnmarshalKey("email.hanqi", &conn)
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
