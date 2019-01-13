package model

import (
	"database/sql"
	"os"
	"strings"
	"time"

	"github.com/guregu/null"
	"github.com/objcoding/wxpay"
	cache "github.com/patrickmn/go-cache"

	"gitlab.com/ftchinese/subscription-api/enum"
	"gitlab.com/ftchinese/subscription-api/paywall"
	"gitlab.com/ftchinese/subscription-api/postoffice"
	"gitlab.com/ftchinese/subscription-api/wechat"

	"github.com/icrowley/fake"
	uuid "github.com/satori/go.uuid"
	"gitlab.com/ftchinese/subscription-api/util"
)

func newDevEnv() Env {
	db, err := sql.Open("mysql", "sampadm:secret@unix(/tmp/mysql.sock)/")

	if err != nil {
		panic(err)
	}

	return Env{
		DB:      db,
		Cache:   cache.New(cache.DefaultExpiration, 0),
		Postman: postoffice.NewPostman(),
	}
}

var appID = os.Getenv("WXPAY_APPID")
var mchID = os.Getenv("WXPAY_MCHID")
var apiKey = os.Getenv("WXPAY_API_KEY")

var mockClient = wechat.NewClient(appID, mchID, apiKey)
var devEnv = newDevEnv()
var mockPlan = paywall.GetDefaultPricing()["standard_year"]
var mockApp = util.ClientApp{
	ClientType: enum.PlatformAndroid,
	Version:    "1.1.1",
	UserIP:     fake.IPv4(),
	UserAgent:  fake.UserAgent(),
}

var tenDaysLater = time.Now().AddDate(0, 0, 10)

type Mocker struct {
	UserID   string
	UnionID  string
	OpenID   string
	UserName string
	Email    string
	IP       string
}

func NewMocker() Mocker {
	unionID, _ := util.RandomBase64(21)
	openID, _ := util.RandomBase64(21)

	return Mocker{
		UserID:   uuid.Must(uuid.NewV4()).String(),
		UnionID:  unionID,
		OpenID:   openID,
		UserName: fake.UserName(),
		Email:    fake.EmailAddress(),
		IP:       fake.IPv4(),
	}
}

func (m Mocker) User() User {
	return User{
		UserID:   m.UserID,
		UnionID:  null.StringFrom(m.UnionID),
		UserName: null.StringFrom(m.UserName),
		Email:    m.Email,
	}
}

func (m Mocker) WxpaySubs() paywall.Subscription {
	return paywall.NewWxpaySubs(m.UserID, mockPlan, enum.EmailLogin)
}

func (m Mocker) AlipaySubs() paywall.Subscription {
	return paywall.NewWxpaySubs(m.UserID, mockPlan, enum.EmailLogin)
}

func (m Mocker) CreateWxpaySubs() (paywall.Subscription, error) {
	subs := paywall.NewWxpaySubs(m.UserID, mockPlan, enum.EmailLogin)

	err := devEnv.SaveSubscription(subs, mockApp)

	if err != nil {
		return subs, err
	}

	return subs, nil
}

func (m Mocker) CreateAlipaySubs() (paywall.Subscription, error) {
	subs := paywall.NewAlipaySubs(m.UserID, mockPlan, enum.EmailLogin)

	err := devEnv.SaveSubscription(subs, mockApp)

	if err != nil {
		return subs, err
	}

	return subs, nil
}

func (m Mocker) CreateMember() (paywall.Subscription, error) {
	subs, err := m.CreateWxpaySubs()

	if err != nil {
		return subs, err
	}

	subs, err = devEnv.ConfirmPayment(subs.OrderID, time.Now())

	if err != nil {
		return subs, err
	}

	return subs, nil
}

func WxNotiResp(orderID string) string {
	openID, _ := util.RandomBase64(21)
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

func WxParsedNoti(orderID string) (wxpay.Params, error) {
	resp := WxNotiResp(orderID)
	return mockClient.ParseResponse(strings.NewReader(resp))
}

func MockPrepay() string {
	p := make(wxpay.Params)

	p = fillResp(p)
	p.SetString("prepay_id", fake.CharactersN(36))

	s := mockClient.Sign(p)

	p.SetString("sign", s)

	return wxpay.MapToXml(p)
}

func MockParsedPrepay() (wxpay.Params, error) {
	resp := MockPrepay()

	return mockClient.ParseResponse(strings.NewReader(resp))
}

func fillResp(p wxpay.Params) wxpay.Params {
	nonce, _ := util.RandomHex(16)

	p.SetString("return_code", "SUCCESS")
	p.SetString("return_msg", "OK")
	p.SetString("appid", appID)
	p.SetString("mch_id", mchID)
	p.SetString("nonce_str", nonce)
	p.SetString("result_code", "SUCCESS")
	p.SetString("trade_type", "APP")

	return p
}

func (u User) createUser() error {
	query := `
	INSERT INTO cmstmp01.userinfo
	SET user_id = ?,
		wx_union_id = ?,
		email = ?,
		password = MD5(?),
		user_name = ?,
		client_type = ?,
		client_version = ?,
		user_ip = INET6_ATON(?),
		user_agent = ?,
		created_utc = UTC_TIMESTAMP()`

	_, err := devEnv.DB.Exec(query,
		u.UserID,
		u.UnionID,
		u.Email,
		fake.Password(8, 20, false, true, false),
		u.UserName,
		mockApp.ClientType,
		mockApp.Version,
		mockApp.UserIP,
		mockApp.UserAgent,
	)
	if err != nil {
		return err
	}
	return nil
}

// Generate a mock subscription that can be used to send a confirmation email.
func (u User) subs() paywall.Subscription {
	subs := paywall.NewWxpaySubs(u.UserID, mockPlan, enum.EmailLogin)
	subs.CreatedAt = util.TimeNow()
	subs.ConfirmedAt = util.TimeNow()
	subs.IsRenewal = false
	subs.StartDate = util.DateNow()
	subs.EndDate = util.DateFrom(time.Now().AddDate(1, 0, 0))

	return subs
}
