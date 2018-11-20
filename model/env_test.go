package model

import (
	"database/sql"

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

var mockPlan = DefaultPlans["standard_year"]

var mockMember = Membership{
	UserID: mockUserID,
	Tier:   TierStandard,
	Cycle:  Yearly,
	Expire: "2018-12-31",
}

var mockClient = util.RequestClient{
	ClientType: "android",
	Version:    "1.1.0",
	UserIP:     fake.IPv4(),
	UserAgent:  "test",
}

var mockSubs = Subscription{
	OrderID:       mockOrderID,
	TierToBuy:     mockPlan.Tier,
	BillingCycle:  mockPlan.Cycle,
	Price:         mockPlan.Price,
	TotalAmount:   mockPlan.Price,
	PaymentMethod: Wxpay,
	CreatedAt:     "2018-11-19T12:12:13Z",
	ConfirmedAt:   "2018-11-19T12:13:13Z",
	IsRenewal:     false,
	StartDate:     "2018-11-19",
	EndDate:       "2019-11-19",
	UserID:        "e1a1f5c0-0e23-11e8-aa75-977ba2bcc6ae",
}

var mockUser = User{
	ID:    mockUserID,
	Name:  "weiguo.ni",
	Email: "weiguo.ni@ftchinese.com",
}
