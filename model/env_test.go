package model

import (
	"database/sql"
	"fmt"
	"github.com/spf13/viper"
	"gitlab.com/ftchinese/subscription-api/util"
	"os"
	"strings"
	"time"

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

func init()  {
	viper.SetConfigName("api")
	viper.AddConfigPath("$HOME/config")
}
// Directly insert a row into ftc_vip for mocking.
func saveMembership(m paywall.Membership) error {
	query := `
	INSERT INTO premium.ftc_vip
	SET vip_id = ?,
		vip_id_alias = ?,
		member_tier = ?,
		billing_cycle = ?,
		expire_date = ?`

	_, err := devEnv.db.Exec(query,
		m.UserID,
		m.UnionID,
		m.Tier,
		m.Cycle,
		m.ExpireDate,
	)

	if err != nil {
		return err
	}

	return nil
}

func newDevDB() *sql.DB {
	db, err := sql.Open("mysql", "sampadm:secret@unix(/tmp/mysql.sock)/")

	if err != nil {
		panic(err)
	}

	return db
}

func newDevPostman() postoffice.Postman {
	//mailHost := os.Getenv("HANQI_SMTP_HOST")
	//mailUser := os.Getenv("HANQI_SMTP_USER")
	//portStr := os.Getenv("HANQI_SMTP_PORT")
	//mailPass := os.Getenv("HANQI_SMTP_PASS")
	//
	//mailPort, _ := strconv.Atoi(portStr)

	var conn util.Conn
	err := viper.UnmarshalKey("hanqi", &conn)
	if err != nil {
		panic(err)
	}
	return postoffice.New(conn.Host, conn.Port, conn.User, conn.Pass)
}

func newWxOAuthApp() wxlogin.WxApp {
	var app wxlogin.WxApp

	err := viper.UnmarshalKey("wxapp.m_subs", &app)
	if err != nil {
		panic(err)
	}

	return app
}

var db = newDevDB()
var postman = newDevPostman()
var devCache = cache.New(cache.DefaultExpiration, 0)
var devEnv = New(db, devCache, false)

var appID = os.Getenv("WXPAY_APPID")
var mchID = os.Getenv("WXPAY_MCHID")
var appSecret = os.Getenv("WXPAY_APPSECRET")
var apiKey = os.Getenv("WXPAY_API_KEY")

var mockClient = wechat.NewClient()

var mockPlan = paywall.GetDefaultPricing()["standard_year"]

func clientApp() gorest.ClientApp {
	return gorest.ClientApp{
		ClientType: enum.PlatformAndroid,
		Version:    "1.1.1",
		UserIP:     fake.IPv4(),
		UserAgent:  fake.UserAgent(),
	}
}

var tenDaysLater = time.Now().AddDate(0, 0, 10)

func generateCode() string {
	code, _ := gorest.RandomBase64(24)
	return code
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
	userID      string
	unionID     string
	email       string
	password    string
	userName    string
	openID      string
	loginMethod enum.LoginMethod
	expireDate  chrono.Date
	ip          string
}

func newMocker() mocker {
	return mocker{
		userID:      uuid.Must(uuid.NewV4()).String(),
		unionID:     generateWxID(),
		email:       fake.EmailAddress(),
		password:    fake.Password(8, 20, false, true, false),
		userName:    fake.UserName(),
		openID:      generateWxID(),
		loginMethod: enum.LoginMethodEmail,
		expireDate:  chrono.DateNow(),
		ip:          fake.IPv4(),
	}
}

func (m mocker) withEmail(email string) mocker {
	m.email = email
	return m
}

func (m mocker) withExpireDate(t time.Time) mocker {
	m.expireDate = chrono.DateFrom(t)
	return m
}

func (m mocker) withWxLogin() mocker {
	m.loginMethod = enum.LoginMethodWx
	return m
}

func (m mocker) user() paywall.User {
	return paywall.User{
		UserID:   m.userID,
		UnionID:  null.StringFrom(m.unionID),
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
		UnionID:      null.StringFrom(m.unionID),
	}
	acc.GenerateSessionID()
	acc.CreatedAt = chrono.TimeNow()
	acc.UpdatedAt = chrono.TimeNow()
	return acc
}

func (m mocker) wxUser() wxlogin.UserInfo {
	return wxlogin.UserInfo{
		UnionID:    m.unionID,
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
	if m.loginMethod == enum.LoginMethodWx {
		return paywall.NewWxpaySubs(m.unionID, mockPlan, enum.LoginMethodWx)
	}
	return paywall.NewWxpaySubs(m.userID, mockPlan, enum.LoginMethodEmail)
}

func (m mocker) alipaySubs() paywall.Subscription {
	if m.loginMethod == enum.LoginMethodWx {
		return paywall.NewAlipaySubs(m.unionID, mockPlan, enum.LoginMethodWx)
	}
	return paywall.NewAlipaySubs(m.userID, mockPlan, enum.LoginMethodEmail)
}

func (m mocker) confirmedSubs() paywall.Subscription {
	subs := paywall.NewWxpaySubs(m.userID, mockPlan, enum.LoginMethodWx)
	subs.CreatedAt = chrono.TimeNow()
	subs.ConfirmedAt = chrono.TimeNow()
	subs.IsRenewal = false
	subs.StartDate = chrono.DateNow()
	subs.EndDate = chrono.DateFrom(time.Now().AddDate(1, 0, 0))

	return subs
}

func (m mocker) member() paywall.Membership {
	mm := paywall.Membership{
		UserID: m.userID,
		Tier:   enum.TierStandard,
		Cycle:  enum.CycleYear,
	}

	if m.loginMethod == enum.LoginMethodWx {
		mm.UnionID = null.StringFrom(m.unionID)
	}

	mm.ExpireDate = m.expireDate

	return mm
}

func (m mocker) createUser() (paywall.User, error) {
	user := m.user()
	app := clientApp()

	query := `
	INSERT INTO cmstmp01.userinfo
	SET user_id = ?,
		email = ?,
		password = MD5(?),
		user_name = ?,
		client_type = ?,
		client_version = ?,
		user_ip = INET6_ATON(?),
		user_agent = ?,
		created_utc = UTC_TIMESTAMP()
	ON DUPLICATE KEY UPDATE
		user_id = ?,
		email = ?,
		password = MD5(?),
		user_name = ?`

	_, err := devEnv.db.Exec(query,
		user.UserID,
		user.Email,
		m.password,
		user.UserName,
		app.ClientType,
		app.Version,
		app.UserIP,
		app.UserAgent,
		user.UserID,
		user.Email,
		m.password,
		user.UserName,
	)
	if err != nil {
		return user, err
	}
	return user, nil
}

func (m mocker) createWxUser() wxlogin.UserInfo {
	userInfo := m.wxUser()

	err := devEnv.SaveWxUser(userInfo)
	if err != nil {
		panic(err)
	}
	return userInfo
}

func (m mocker) createWxpaySubs() paywall.Subscription {
	subs := m.wxpaySubs()

	err := devEnv.SaveSubscription(subs, clientApp())

	if err != nil {
		panic(err)
	}

	return subs
}

func (m mocker) createAlipaySubs() paywall.Subscription {
	subs := m.alipaySubs()

	err := devEnv.SaveSubscription(subs, clientApp())

	if err != nil {
		panic(err)
	}

	return subs
}

func (m mocker) createMember() paywall.Membership {
	mm := m.member()
	err := saveMembership(mm)

	if err != nil {
		panic(err)
	}

	return mm
}

func (m mocker) createWxAccess() wxlogin.OAuthAccess {
	acc := m.wxAccess()

	err := devEnv.SaveWxAccess(appID, acc, clientApp())

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
	p.SetString("appid", appID)
	p.SetString("mch_id", mchID)
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
