package model

import (
	"testing"
	"time"

	"gitlab.com/ftchinese/subscription-api/util"
)

var mockPlan = plans["standard_year"]

const mockOrderID = "FT0102381539932302"

var mockSubs = Subscription{
	OrderID:       mockOrderID,
	TierToBuy:     mockPlan.Tier,
	BillingCycle:  mockPlan.Cycle,
	Price:         float32(mockPlan.Price),
	TotalAmount:   float32(mockPlan.Price),
	PaymentMethod: Wxpay,
	UserID:        "e1a1f5c0-0e23-11e8-aa75-977ba2bcc6ae",
}

func TestNewSubscription(t *testing.T) {

	c := util.RequestClient{
		ClientType: "android",
		Version:    "0.0.1",
		UserIP:     "127.0.0.1",
	}

	err := devEnv.NewSubscription(mockSubs, c)

	if err != nil {
		t.Error(err)
	}
}

func TestRetrieveSubscription(t *testing.T) {
	s, err := devEnv.FindSubscription("FT0102661537944423")

	if err != nil {
		t.Error(err)
	}

	t.Logf("%+v\n", s)
}

func TestWxTotalFee(t *testing.T) {
	t.Log(int64(198.00*100) == 19800)
}

func TestConfirmSubscription(t *testing.T) {
	err := devEnv.ConfirmSubscription(mockSubs, time.Now())

	if err != nil {
		t.Error(err)
	}
}
