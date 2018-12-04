package model

import (
	"testing"
	"time"
)

// Test Subscription's withConfirmation method.
func TestSubsConfirm(t *testing.T) {

	subs, err := confirmSubs(false)

	if err != nil {
		t.Error(err)
	}

	t.Logf("Subscritpion confirmed: %+v\n", subs)
}

// Test Subscription's withMembership method.
func TestSubsRenew(t *testing.T) {

	subs, err := confirmSubs(true)

	if err != nil {
		t.Error(err)

		return
	}

	t.Logf("Renw membership with subscription: %+v\n", subs)
}

// Test create a new subscription order
func TestSaveSubs(t *testing.T) {

	subs, err := insertSubs(false)

	if err != nil {
		t.Error(err)
	}

	t.Log(subs)
}

func TestRetrieveSubs(t *testing.T) {
	subs, err := createAndFindSubs(false)

	if err != nil {
		t.Error(err)
	}

	t.Logf("%+v\n", subs)
}

func TestConfirmSubs__new(t *testing.T) {
	// Create a new subscription order and retrieve it.
	subs, err := createAndFindSubs(false)

	if err != nil {
		t.Error(err)
		return
	}

	subs, err = devEnv.ConfirmSubscription(subs, time.Now())

	if err != nil {
		t.Error(err)
	}

	t.Logf("%+v\n", subs)
}

func TestConfirmSubs__renew(t *testing.T) {
	subs, err := createAndFindSubs(true)

	if err != nil {
		t.Error(err)
		return
	}

	subs, err = devEnv.ConfirmSubscription(subs, time.Now())

	if err != nil {
		t.Error(err)
	}

	t.Logf("%+v\n", subs)
}

// Test create a new member.
// Workflow as follows:
// Client
func TestCreateMember(t *testing.T) {
	subs, err := createAndFindSubs(false)

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

	t.Logf("%+v\n", subs)
}

func TestRenewMember(t *testing.T) {
	subs, err := createAndFindSubs(true)

	subs, err = devEnv.ConfirmSubscription(subs, time.Now())

	if err != nil {
		t.Error(err)
		return
	}

	err = devEnv.CreateOrUpdateMember(subs)

	if err != nil {
		t.Error(err)
	}

	t.Log(subs)
}
