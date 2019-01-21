package model

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/postoffice"
	"github.com/smartwalle/alipay"

	randomdata "github.com/Pallinder/go-randomdata"
	"gitlab.com/ftchinese/subscription-api/wxlogin"

	"github.com/guregu/null"
	"github.com/objcoding/wxpay"
	cache "github.com/patrickmn/go-cache"

	"github.com/FTChinese/go-rest/enum"
	"gitlab.com/ftchinese/subscription-api/paywall"

	"gitlab.com/ftchinese/subscription-api/wechat"

	"github.com/icrowley/fake"
	uuid "github.com/satori/go.uuid"
)

func newDevDB() *sql.DB {
	db, err := sql.Open("mysql", "sampadm:secret@unix(/tmp/mysql.sock)/")

	if err != nil {
		panic(err)
	}

	return db
}

func newDevPostman() postoffice.Postman {
	mailHost := os.Getenv("HANQI_SMTP_HOST")
	mailUser := os.Getenv("HANQI_SMTP_USER")
	portStr := os.Getenv("HANQI_SMTP_PORT")
	mailPass := os.Getenv("HANQI_SMTP_PASS")

	mailPort, _ := strconv.Atoi(portStr)

	return postoffice.NewPostman(mailHost, mailPort, mailUser, mailPass)
}

var db = newDevDB()
var postman = newDevPostman()

func newDevEnv() Env {
	db, err := sql.Open("mysql", "sampadm:secret@unix(/tmp/mysql.sock)/")

	if err != nil {
		panic(err)
	}

	return Env{
		db:    db,
		cache: cache.New(cache.DefaultExpiration, 0),
	}
}

var appID = os.Getenv("WXPAY_APPID")
var mchID = os.Getenv("WXPAY_MCHID")
var appSecret = os.Getenv("WXPAY_APPSECRET")
var apiKey = os.Getenv("WXPAY_API_KEY")

var mockClient = wechat.NewClient(appID, mchID, apiKey)
var devEnv = newDevEnv()
var mockPlan = paywall.GetDefaultPricing()["standard_year"]
var mockApp = gorest.ClientApp{
	ClientType: enum.PlatformAndroid,
	Version:    "1.1.1",
	UserIP:     fake.IPv4(),
	UserAgent:  fake.UserAgent(),
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
	openID      string
	loginMethod enum.LoginMethod
	userName    string
	email       string
	password    string
	ip          string
}

func newMocker() mocker {
	return mocker{
		userID:      uuid.Must(uuid.NewV4()).String(),
		unionID:     generateWxID(),
		openID:      generateWxID(),
		loginMethod: enum.LoginMethodEmail,
		userName:    fake.UserName(),
		email:       fake.EmailAddress(),
		password:    fake.Password(8, 20, false, true, false),
		ip:          fake.IPv4(),
	}
}

func (m mocker) withEmail(email string) mocker {
	m.email = email
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
	return paywall.NewWxpaySubs(m.userID, mockPlan, enum.LoginMethodWx)
}

func (m mocker) alipaySubs() paywall.Subscription {
	if m.loginMethod == enum.LoginMethodWx {
		return paywall.NewAlipaySubs(m.unionID, mockPlan, enum.LoginMethodWx)
	}
	return paywall.NewAlipaySubs(m.userID, mockPlan, enum.LoginMethodWx)
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

func (m mocker) createUser() (paywall.User, error) {
	user := m.user()

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
		mockApp.ClientType,
		mockApp.Version,
		mockApp.UserIP,
		mockApp.UserAgent,
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

func (m mocker) createWxUser() (wxlogin.UserInfo, error) {
	userInfo := m.wxUser()

	err := devEnv.SaveWxUser(userInfo)
	if err != nil {
		return userInfo, err
	}
	return userInfo, nil
}

func (m mocker) createWxpaySubs() (paywall.Subscription, error) {
	subs := m.wxpaySubs()

	err := devEnv.SaveSubscription(subs, mockApp)

	if err != nil {
		return subs, err
	}

	return subs, nil
}

func (m mocker) createAlipaySubs() (paywall.Subscription, error) {
	subs := m.alipaySubs()

	err := devEnv.SaveSubscription(subs, mockApp)

	if err != nil {
		return subs, err
	}

	return subs, nil
}

func (m mocker) createMember() (paywall.Subscription, error) {
	subs, err := m.createWxpaySubs()

	if err != nil {
		return subs, err
	}

	subs, err = devEnv.ConfirmPayment(subs.OrderID, time.Now())

	if err != nil {
		return subs, err
	}

	return subs, nil
}

func (m mocker) createWxAccess() (wxlogin.OAuthAccess, error) {
	acc := m.wxAccess()

	err := devEnv.SaveWxAccess(appID, acc, mockApp)

	if err != nil {
		return acc, err
	}

	return acc, nil
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

func wxParsedNoti(orderID string) (wxpay.Params, error) {
	resp := wxNotiResp(orderID)
	return mockClient.ParseResponse(strings.NewReader(resp))
}

func wxPrepayResp() string {
	p := make(wxpay.Params)

	p = fillResp(p)
	p.SetString("prepay_id", fake.CharactersN(36))

	s := mockClient.Sign(p)

	p.SetString("sign", s)

	return wxpay.MapToXml(p)
}

func wxParsedPrepay() (wxpay.Params, error) {
	resp := wxPrepayResp()

	return mockClient.ParseResponse(strings.NewReader(resp))
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
