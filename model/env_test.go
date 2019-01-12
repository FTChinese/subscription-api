package model

import (
	"database/sql"
	"testing"
	"time"

	"github.com/guregu/null"

	"gitlab.com/ftchinese/subscription-api/enum"
	"gitlab.com/ftchinese/subscription-api/paywall"
	"gitlab.com/ftchinese/subscription-api/postoffice"

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
		Postman: postoffice.NewPostman(),
	}
}

var devEnv = newDevEnv()
var mockPlan = paywall.GetDefaultPricing()["standard_year"]
var mockClient = util.ClientApp{
	ClientType: enum.PlatformAndroid,
	Version:    "1.1.1",
	UserIP:     fake.IPv4(),
	UserAgent:  fake.UserAgent(),
}

var tenDaysLater = time.Now().AddDate(0, 0, 10)

func NewUser() User {
	unionID, _ := util.RandomBase64(21)

	return User{
		UserID:   uuid.Must(uuid.NewV4()).String(),
		UnionID:  null.StringFrom(unionID),
		UserName: null.StringFrom(fake.UserName()),
		Email:    fake.EmailAddress(),
	}
}

func (u User) subs() paywall.Subscription {
	subs := paywall.NewWxpaySubs(u.UserID, mockPlan, enum.EmailLogin)
	subs.CreatedAt = util.TimeNow()
	subs.ConfirmedAt = util.TimeNow()
	subs.IsRenewal = false
	subs.StartDate = util.DateNow()
	subs.EndDate = util.DateFrom(time.Now().AddDate(1, 0, 0))

	return subs
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
		mockClient.ClientType,
		mockClient.Version,
		mockClient.UserIP,
		mockClient.UserAgent,
	)
	if err != nil {
		return err
	}
	return nil
}

func (u User) CreateWxpaySubs() (paywall.Subscription, error) {
	subs := paywall.NewWxpaySubs(u.UserID, mockPlan, enum.EmailLogin)

	err := devEnv.SaveSubscription(subs, mockClient)

	if err != nil {
		return subs, err
	}

	return subs, nil
}

func (u User) CreateAlipaySubs() (paywall.Subscription, error) {
	subs := paywall.NewAlipaySubs(u.UserID, mockPlan, enum.EmailLogin)

	err := devEnv.SaveSubscription(subs, mockClient)

	if err != nil {
		return subs, err
	}

	return subs, nil
}

func (u User) CreateMember() (paywall.Subscription, error) {
	subs, err := u.CreateWxpaySubs()

	if err != nil {
		return subs, err
	}

	subs, err = devEnv.ConfirmPayment(subs.OrderID, time.Now())

	if err != nil {
		return subs, err
	}

	return subs, nil
}

func TestDBConn(t *testing.T) {
	err := devEnv.DB.Ping()

	if err != nil {
		t.Error(err)
	}
}
