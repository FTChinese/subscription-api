package test

import (
	"database/sql"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/postoffice"
	"github.com/Pallinder/go-randomdata"
	"github.com/guregu/null"
	"github.com/patrickmn/go-cache"
	"github.com/spf13/viper"
	"gitlab.com/ftchinese/subscription-api/paywall"
	"gitlab.com/ftchinese/subscription-api/util"
	"gitlab.com/ftchinese/subscription-api/wechat"
	"gitlab.com/ftchinese/subscription-api/wxlogin"
)

var DB *sql.DB
var Postman postoffice.Postman
var Cache *cache.Cache
var WxOauthApp wxlogin.WxApp
var WxPayApp wechat.PayApp
var WxPayClient wechat.Client

var TestPlan = paywall.GetDefaultPricing()["standard_year"]

func init() {
	viper.SetConfigName("api")
	viper.AddConfigPath("$HOME/config")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	DB, err = sql.Open("mysql", "sampadm:secret@unix(/tmp/mysql.sock)/")
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

	err = viper.UnmarshalKey("wxapp.m_subs", &WxOauthApp)
	if err != nil {
		panic(err)
	}

	err = viper.UnmarshalKey("wxapp.m_subs", &WxPayApp)
	if err != nil {
		panic(err)
	}

	WxPayClient = wechat.NewClient(WxPayApp)
}

func RandomClientApp() util.ClientApp {
	return util.ClientApp{
		ClientType: enum.PlatformAndroid,
		Version:    null.StringFrom("1.1.1"),
		UserIP:     null.StringFrom(randomdata.IpV4Address()),
		UserAgent:  null.StringFrom(randomdata.UserAgentString()),
	}
}
