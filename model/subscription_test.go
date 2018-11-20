package model

import (
	"testing"
	"time"

	"gitlab.com/ftchinese/subscription-api/util"
)

func TestSaveSubscription(t *testing.T) {

	c := util.RequestClient{
		ClientType: "android",
		Version:    "0.0.1",
		UserIP:     "127.0.0.1",
	}

	err := devEnv.SaveSubscription(mockSubs, c)

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

func TestSubsConfirm(t *testing.T) {
	now := time.Now()

	subs := mockSubs.confirm(now)

	t.Logf("Subscritpion confirmed: %+v\n", subs)
}

func TestSubsRenew(t *testing.T) {

	mockSubs.IsRenewal = true
	subs, err := mockSubs.renew(mockMember)

	if err != nil {
		t.Error(err)

		return
	}

	t.Logf("Renw membership with subscription: %+v\n", subs)
}

func TestNewSubscription(t *testing.T) {
	// 1. Get a plan
	plan, ok := DefaultPlans["standard_year"]

	if !ok {
		t.Error("No plan found")
		return
	}

	// 2. Create a subscription order based on the plan
	subs := plan.CreateOrder(mockUserID, Wxpay)

	// 3. Find out if this membership is for renewal of a new one.
	member, err := devEnv.FindMember(mockUserID)

	if err == nil {
		subs.IsRenewal = !member.IsExpired()
	}

	// 4. Save order
	err = devEnv.SaveSubscription(subs, mockClient)

	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Saved subscription order: %+v\n", subs)

	// 5. Confirm order is paid
	updatedSubs, err := devEnv.ConfirmSubscription(subs, time.Now())

	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Confirmed subscription: %+v\n", updatedSubs)

	// Create or update a membership based on whether it exists.
	err = devEnv.CreateOrUpdateMember(updatedSubs)

	if err != nil {
		t.Error(err)
		return
	}
}

func TestUpdateSubscription(t *testing.T) {
	err := devEnv.CreateOrUpdateMember(mockSubs)

	if err != nil {
		t.Error(err)
	}
}
