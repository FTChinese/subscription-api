package model

import (
	"database/sql"
	"fmt"
	"github.com/pkg/errors"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
	"gitlab.com/ftchinese/subscription-api/util"

	"github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/postoffice"
	"github.com/smartwalle/alipay"

	"github.com/Pallinder/go-randomdata"
	"gitlab.com/ftchinese/subscription-api/wxlogin"

	"github.com/guregu/null"
	"github.com/objcoding/wxpay"
	"github.com/patrickmn/go-cache"

	"github.com/FTChinese/go-rest/enum"
	"gitlab.com/ftchinese/subscription-api/paywall"

	"gitlab.com/ftchinese/subscription-api/wechat"

	"github.com/icrowley/fake"
	"github.com/satori/go.uuid"

	_ "github.com/go-sql-driver/mysql"
)

const (
	myFtcID    = "e1a1f5c0-0e23-11e8-aa75-977ba2bcc6ae"
	myFtcEmail = "neefrankie@163.com"
	myUnionID  = "ogfvwjk6bFqv2yQpOrac0J3PqA0o"
)

var db *sql.DB
var postman postoffice.Postman
var devCache *cache.Cache
var oauthApp wxlogin.WxApp
var wxpayApp wechat.PayApp
var mockClient wechat.Client

var mockPlan = paywall.GetDefaultPricing()["standard_year"]

func init() {
	viper.SetConfigName("api")
	viper.AddConfigPath("$HOME/config")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	db, err = sql.Open("mysql", "sampadm:secret@unix(/tmp/mysql.sock)/")
	if err != nil {
		panic(err)
	}

	var conn util.Conn
	err = viper.UnmarshalKey("hanqi", &conn)
	if err != nil {
		panic(err)
	}
	postman = postoffice.New(conn.Host, conn.Port, conn.User, conn.Pass)

	devCache = cache.New(cache.DefaultExpiration, 0)

	err = viper.UnmarshalKey("wxapp.m_subs", &oauthApp)
	if err != nil {
		panic(err)
	}

	err = viper.UnmarshalKey("wxapp.m_subs", &wxpayApp)
	if err != nil {
		panic(err)
	}

	mockClient = wechat.NewClient(wxpayApp)
}

func clientApp() gorest.ClientApp {
	return gorest.ClientApp{
		ClientType: enum.PlatformAndroid,
		Version:    "1.1.1",
		UserIP:     fake.IPv4(),
		UserAgent:  fake.UserAgent(),
	}
}

func generateCode() string {
	code, _ := gorest.RandomBase64(24)
	return code
}

func genUUID() string {
	return uuid.Must(uuid.NewV4()).String()
}

func generateToken() string {
	token, _ := gorest.RandomBase64(82)
	return token
}

func generateWxID() string {
	id, _ := gorest.RandomBase64(21)
	return id
}

func generateAvatarURL() string {
	return fmt.Sprintf("http://thirdwx.qlogo.cn/mmopen/vi_32/%s/132", fake.CharactersN(90))
}

type mocker struct {
	userID     null.String
	unionID    null.String
	email      string
	password   string
	userName   string
	openID     string
	expireDate chrono.Date
	ip         string
}

func newMocker() mocker {
	return mocker{
		email:      fake.EmailAddress(),
		password:   fake.Password(8, 20, false, true, false),
		userName:   fake.UserName(),
		openID:     generateWxID(),
		expireDate: chrono.DateNow(),
		ip:         fake.IPv4(),
	}
}

func (m mocker) withUserID() mocker {
	m.userID = null.StringFrom(uuid.Must(uuid.NewV4()).String())
	return m
}

func (m mocker) withUnionID() mocker {
	m.unionID = null.StringFrom(generateWxID())
	return m
}

func (m mocker) bound() mocker {
	m.userID = null.StringFrom(uuid.Must(uuid.NewV4()).String())
	m.unionID = null.StringFrom(generateWxID())

	return m
}

func (m mocker) withExpireDate(t time.Time) mocker {
	m.expireDate = chrono.DateFrom(t)
	return m
}

func (m mocker) compoundID() string {
	if m.userID.Valid {
		return m.userID.String
	} else if m.unionID.Valid {
		return m.unionID.String
	} else {
		panic(errors.New("Both user id and union id are empty"))
	}
}

func (m mocker) user() paywall.User {
	return paywall.User{
		UserID:   m.userID.String,
		UnionID:  m.unionID,
		UserName: null.StringFrom(m.userName),
		Email:    m.email,
	}
}

func (m mocker) wxAccess() wxlogin.OAuthAccess {
	acc := wxlogin.OAuthAccess{
		AccessToken:  generateToken(),
		ExpiresIn:    7200,
		RefreshToken: generateToken(),
		OpenID:       m.openID,
		Scope:        "snsapi_userinfo",
		UnionID:      m.unionID,
	}
	acc.GenerateSessionID()
	acc.CreatedAt = chrono.TimeNow()
	acc.UpdatedAt = chrono.TimeNow()
	return acc
}

func (m mocker) wxUserInfo() wxlogin.UserInfo {
	return wxlogin.UserInfo{
		UnionID:    m.unionID.String,
		NickName:   fake.UserName(),
		AvatarURL:  generateAvatarURL(),
		Sex:        randomdata.Number(0, 3),
		Country:    fake.Country(),
		Province:   fake.State(),
		City:       fake.City(),
		Privileges: []string{},
	}
}

func (m mocker) wxpaySubs() paywall.Subscription {
	s, _ := paywall.NewWxpaySubs(m.userID, m.unionID, mockPlan)
	return s
}

func (m mocker) alipaySubs() paywall.Subscription {
	s, _ := paywall.NewAlipaySubs(m.userID, m.unionID, mockPlan)
	return s
}

func (m mocker) confirmedSubs() paywall.Subscription {
	subs, err := paywall.NewWxpaySubs(m.userID, m.unionID, mockPlan)
	if err != nil {
		panic(err)
	}
	subs.CreatedAt = chrono.TimeNow()
	subs.ConfirmedAt = chrono.TimeNow()
	subs.IsRenewal = false
	subs.StartDate = chrono.DateNow()
	subs.EndDate = chrono.DateFrom(time.Now().AddDate(1, 0, 0))

	return subs
}

func (m mocker) member() paywall.Membership {
	mm := paywall.Membership{
		CompoundID: m.compoundID(),
		UnionID:    m.unionID,
		Tier:       enum.TierStandard,
		Cycle:      enum.CycleYear,
	}

	mm.ExpireDate = m.expireDate

	return mm
}

func (m mocker) createWxUser() wxlogin.UserInfo {
	userInfo := m.wxUserInfo()

	env := Env{db: db}
	err := env.SaveWxUser(userInfo)
	if err != nil {
		panic(err)
	}
	return userInfo
}

func (m mocker) createWxpaySubs() paywall.Subscription {
	subs := m.wxpaySubs()

	env := Env{db: db}
	err := env.SaveSubscription(subs, clientApp())

	if err != nil {
		panic(err)
	}

	return subs
}

func (m mocker) createAlipaySubs() paywall.Subscription {
	subs := m.alipaySubs()

	env := Env{db: db}
	err := env.SaveSubscription(subs, clientApp())

	if err != nil {
		panic(err)
	}

	return subs
}

func (m mocker) createMember() paywall.Membership {

	mm := paywall.Membership{
		CompoundID: m.compoundID(),
		UnionID:    m.unionID,
		Tier:       enum.TierStandard,
		Cycle:      enum.CycleYear,
	}

	env := Env{db: db}
	_, err := db.Exec(env.stmtInsertMember(),
		mm.CompoundID,
		mm.UnionID,
		m.userID,
		mm.UnionID,
		mm.Tier,
		mm.Cycle,
		m.expireDate,
	)

	if err != nil {
		panic(err)
	}

	return mm
}

func (m mocker) createWxAccess() wxlogin.OAuthAccess {
	acc := m.wxAccess()

	env := Env{db: db}
	err := env.SaveWxAccess(oauthApp.AppID, acc, clientApp())

	if err != nil {
		panic(err)
	}

	return acc
}

func wxNotiResp(orderID string) string {
	openID, _ := gorest.RandomBase64(21)
	p := make(wxpay.Params)

	p = fillResp(p)
	p.SetString("openid", openID)
	p.SetString("is_subscribe", "N")
	p.SetString("bank_type", "CMC")
	p.SetString("total_fee", "25800")
	p.SetString("cash_fee", "25800")
	p.SetString("transaction_id", fake.CharactersN(28))
	p.SetString("out_trade_no", orderID)
	p.SetString("time_end", time.Now().Format("20060102150405"))

	s := mockClient.Sign(p)

	p.SetString("sign", s)

	return wxpay.MapToXml(p)
}

func wxPrepayResp() string {
	p := make(wxpay.Params)

	p = fillResp(p)
	p.SetString("prepay_id", fake.CharactersN(36))

	s := mockClient.Sign(p)

	p.SetString("sign", s)

	return wxpay.MapToXml(p)
}

func wxParsedPrepay() wxpay.Params {
	resp := wxPrepayResp()

	p, err := mockClient.ParseResponse(strings.NewReader(resp))

	if err != nil {
		panic(err)
	}

	return p
}

func wxParsedNoti(orderID string) wxpay.Params {
	resp := wxNotiResp(orderID)
	p, err := mockClient.ParseResponse(strings.NewReader(resp))
	if err != nil {
		panic(err)
	}

	return p
}

func fillResp(p wxpay.Params) wxpay.Params {
	nonce, _ := gorest.RandomHex(16)

	p.SetString("return_code", "SUCCESS")
	p.SetString("return_msg", "OK")
	p.SetString("appid", wxpayApp.AppID)
	p.SetString("mch_id", wxpayApp.MchID)
	p.SetString("nonce_str", nonce)
	p.SetString("result_code", "SUCCESS")
	p.SetString("trade_type", "APP")

	return p
}

func aliNoti() alipay.TradeNotification {
	return alipay.TradeNotification{
		NotifyTime: time.Now().In(time.UTC).Format(chrono.SQLDateTime),
		NotifyType: "trade_status_sync",
		NotifyId:   fake.CharactersN(36),
		AppId:      os.Getenv("ALIPAY_APP_ID"),
		Charset:    "utf-8",
		Version:    "1.0",
		SignType:   "RSA2",
		Sign:       fake.CharactersN(256),
		TradeNo:    fake.CharactersN(64),
		OutTradeNo: fake.CharactersN(18),
		GmtCreate:  time.Now().In(time.UTC).Format(chrono.SQLDateTime),
		GmtPayment: time.Now().In(time.UTC).Format(chrono.SQLDateTime),
	}
}
