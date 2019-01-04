package model

import (
	"database/sql"
	"time"

	"gitlab.com/ftchinese/subscription-api/enum"

	"github.com/icrowley/fake"
	cache "github.com/patrickmn/go-cache"
	uuid "github.com/satori/go.uuid"
	"gitlab.com/ftchinese/subscription-api/util"
)

func newDevEnv() Env {
	db, err := sql.Open("mysql", "sampadm:secret@unix(/tmp/mysql.sock)/")

	if err != nil {
		panic(err)
	}

	c := cache.New(cache.DefaultExpiration, 0)

	return Env{DB: db, Cache: c}
}

var devEnv = newDevEnv()

var mockUUID = uuid.Must(uuid.NewV4()).String()

var mockUnionID, _ = util.RandomBase64(21)

var mockClient = util.ClientApp{
	ClientType: enum.PlatformAndroid,
	Version:    "1.1.1",
	UserIP:     fake.IPv4(),
	UserAgent:  fake.UserAgent(),
}

var mockPlan = DefaultPlans["standard_year"]

var tenDaysLater = time.Now().AddDate(0, 0, 10)

// Mock inserting a subscription order.
// isRenew dtermines is this order is used to
// renew a membership or not.
func createSubs(isWxLogin bool) (Subscription, error) {
	var subs Subscription
	if isWxLogin {
		subs = NewWxSubs(mockUnionID, mockPlan, enum.WechatLogin)
	} else {
		subs = NewWxSubs(mockUUID, mockPlan, enum.EmailLogin)
	}

	err := devEnv.SaveSubscription(subs, mockClient)

	if err != nil {
		return subs, err
	}

	return subs, nil
}

// Mock creating a new membership.
// Return user id so that it could be used to generate a new subscription order for the same user.
func createMember(isWxLogin bool) (Subscription, error) {
	subs, err := createSubs(isWxLogin)

	if err != nil {
		return subs, err
	}

	subs, err = devEnv.ConfirmSubscription(subs, time.Now())

	if err != nil {
		return subs, err
	}

	err = devEnv.CreateMembership(subs)

	if err != nil {
		return subs, err
	}

	return subs, nil
}
