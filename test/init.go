package test

import (
	"database/sql"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/postoffice"
	"github.com/patrickmn/go-cache"
	"github.com/spf13/viper"
	"gitlab.com/ftchinese/subscription-api/models/paywall"
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

var YearlyStandard = paywall.GetFtcPlans(false)["standard_year"]
var YearlyPremium = paywall.GetFtcPlans(false)["premium_year"]

var (
	DB          *sql.DB
	Postman     postoffice.Postman
	Cache       *cache.Cache
	WxOAuthApp  wxlogin.WxApp
	WxPayApp    wechat.PayApp
	WxPayClient wechat.Client
	StripeKey   string

	PlanStandardMonthly = paywall.Plan{
		Coordinate: paywall.Coordinate{
			Tier:  enum.TierStandard,
			Cycle: enum.CycleMonth,
		},
		ListPrice:  28.00,
		NetPrice:   28.00,
		Title:      "FT中文网 - 月度标准会员",
		CycleCount: 1,
		Currency:   "cny",
		ExtraDays:  1,
		StripeID:   "plan_FOdgPTznDwHU4i",
	}
	PlanStandardYearly = paywall.Plan{
		Coordinate: paywall.Coordinate{
			Tier:  enum.TierStandard,
			Cycle: enum.CycleYear,
		},
		ListPrice:  258.00,
		NetPrice:   258.00,
		Title:      "FT中文网 - 年度标准会员",
		CycleCount: 1,
		Currency:   "cny",
		ExtraDays:  1,
		StripeID:   "plan_FOdfeaqzczp6Ag",
	}
	PlanPremiumYearly = paywall.Plan{
		Coordinate: paywall.Coordinate{
			Tier:  enum.TierPremium,
			Cycle: enum.CycleYear,
		},
		ListPrice:  1998.00,
		NetPrice:   1998.00,
		Title:      "FT中文网 - 高端会员",
		CycleCount: 1,
		Currency:   "cny",
		ExtraDays:  1,
		StripeID:   "plan_FOde0uAr0V4WmT",
	}
)

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
