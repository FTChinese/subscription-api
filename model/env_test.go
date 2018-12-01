package model

import (
	"database/sql"
	"time"

	"gitlab.com/ftchinese/subscription-api/member"

	"github.com/icrowley/fake"
	"gitlab.com/ftchinese/subscription-api/util"

	cache "github.com/patrickmn/go-cache"
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

const mockOrderID = "FT0102381539932302"
const mockUserID = "e1a1f5c0-0e23-11e8-aa75-977ba2bcc6ae"

var mockClient = util.RequestClient{
	ClientType: "android",
	Version:    "1.1.1",
	UserIP:     fake.IPv4(),
	UserAgent:  "golang mocker",
}

var mockPlan = DefaultPlans["standard_year"]

var tommorrow = util.SQLDateUTC.FromTime(time.Now().AddDate(0, 0, 1))

var mockMember = Membership{
	UserID: mockUserID,
	Tier:   member.TierStandard,
	Cycle:  member.CycleYear,
	Expire: tommorrow,
}

var mockUser = User{
	ID:    mockUserID,
	Name:  "weiguo.ni",
	Email: "weiguo.ni@ftchinese.com",
}

func insertSubs(isRenew bool) (Subscription, error) {
	subs := mockPlan.CreateSubs(mockUserID, member.Wxpay)

	subs.IsRenewal = isRenew

	err := devEnv.SaveSubscription(subs, mockClient)

	if err != nil {
		return subs, err
	}

	return subs, nil
}

func confirmSubs() (Subscription, error) {
	subs, err := insertSubs(false)

	if err != nil {
		return subs, err
	}

	now := time.Now()

	subs, err = subs.withConfirmation(now)

	if err != nil {
		return subs, err
	}

	return subs, nil
}

func renewSubs() (Subscription, error) {
	subs, err := insertSubs(true)

	if err != nil {
		return subs, err
	}

	now := time.Now()

	subs, err = subs.withConfirmation(now)

	if err != nil {
		return subs, err
	}

	subs, err = subs.withMembership(mockMember)

	if err != nil {
		return subs, err
	}

	return subs, nil
}
