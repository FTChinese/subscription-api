package test

import (
	"github.com/FTChinese/go-rest/connect"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/postoffice"
	"github.com/jmoiron/sqlx"
	"github.com/patrickmn/go-cache"
	"github.com/spf13/viper"
	"gitlab.com/ftchinese/subscription-api/models/plan"
	"gitlab.com/ftchinese/subscription-api/models/wechat"
	"gitlab.com/ftchinese/subscription-api/models/wxlogin"
	"gitlab.com/ftchinese/subscription-api/pkg/config"
	"gitlab.com/ftchinese/subscription-api/pkg/db"
	"log"
)

const (
	MyFtcID    = "e1a1f5c0-0e23-11e8-aa75-977ba2bcc6ae"
	MyFtcEmail = "neefrankie@163.com"
	MyUnionID  = "ogfvwjk6bFqv2yQpOrac0J3PqA0o"
	MyEmail    = "neefrankie@gmail.com"
)

func mustFindPlan(tier enum.Tier, cycle enum.Cycle) plan.Plan {
	p, err := plan.FindPlan(tier, cycle)
	if err != nil {
		panic(err)
	}

	return p
}

var YearlyStandard = mustFindPlan(enum.TierStandard, enum.CycleYear)
var YearlyPremium = mustFindPlan(enum.TierPremium, enum.CycleYear)

var (
	DB          *sqlx.DB
	Postman     postoffice.PostOffice
	Cache       *cache.Cache
	WxOAuthApp  wxlogin.WxApp
	WxPayApp    wechat.PayApp
	WxPayClient wechat.Client
	StripeKey   string
)

func init() {
	viper.SetConfigName("api")
	viper.AddConfigPath("$HOME/config")

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal(err)
	}

	var dbConn connect.Connect
	err = viper.UnmarshalKey("mysql.dev", &dbConn)
	if err != nil {
		panic(err)
	}

	DB = db.MustNewDB(dbConn)

	Postman = postoffice.New(config.MustGetHanqiConn())

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
