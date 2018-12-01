package model

import (
	"testing"
	"time"
)

func TestSaveSubs(t *testing.T) {

	subs, err := insertSubs(false)

	if err != nil {
		t.Error(err)
	}

	t.Log(subs)
}

func TestRetrieveSubs(t *testing.T) {
	subs, err := insertSubs(false)

	s, err := devEnv.FindSubscription(subs.OrderID)

	if err != nil {
		t.Error(err)
	}

	t.Logf("%+v\n", s)
}

func TestWxTotalFee(t *testing.T) {
	t.Log(int64(198.00*100) == 19800)
}

func TestSubsConfirm(t *testing.T) {

	subs, err := confirmSubs()

	if err != nil {
		t.Error(err)
	}

	t.Logf("Subscritpion confirmed: %+v\n", subs)
}

func TestSubsRenew(t *testing.T) {

	subs, err := renewSubs()

	if err != nil {
		t.Error(err)

		return
	}

	t.Logf("Renw membership with subscription: %+v\n", subs)
}

func TestConfirmSubs__new(t *testing.T) {
	// Create a new subscription order.
	subs, err := insertSubs(false)

	subs, err = devEnv.ConfirmSubscription(subs, time.Now())

	if err != nil {
		t.Error(err)
	}

	t.Logf("%+v\n", subs)
}

func TestConfirmSubs__renew(t *testing.T) {
	subs, err := insertSubs(true)

	subs, err = devEnv.ConfirmSubscription(subs, time.Now())

	if err != nil {
		t.Error(err)
	}

	t.Logf("%+v\n", subs)
}
func TestCreateMember(t *testing.T) {
	subs, err := insertSubs(false)

	subs, err = devEnv.ConfirmSubscription(subs, time.Now())

	if err != nil {
		t.Error(err)
		return
	}

	err = devEnv.CreateOrUpdateMember(subs)

	if err != nil {
		t.Error(err)
		return
	}

	t.Log(subs)
}

func TestRenewMember(t *testing.T) {
	subs, err := insertSubs(true)

	subs, err = devEnv.ConfirmSubscription(subs, time.Now())

	if err != nil {
		t.Error(err)
		return
	}

	err = devEnv.CreateOrUpdateMember(subs)

	if err != nil {
		t.Error(err)
		return
	}

	t.Log(subs)
}
