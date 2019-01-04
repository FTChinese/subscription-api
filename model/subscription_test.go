package model

import (
	"testing"
	"time"

	"gitlab.com/ftchinese/subscription-api/enum"
	"gitlab.com/ftchinese/subscription-api/util"
)

func TestNewFTCSubs(t *testing.T) {
	wxSubs := NewWxSubs(mockUUID, mockPlan, enum.EmailLogin)

	t.Logf("Created a Wechat subscription: %+v\n", wxSubs)

	aliSubs := NewAliSubs(mockUUID, mockPlan, enum.EmailLogin)

	t.Logf("Created a Ali subscription: %+v\n", aliSubs)
}

func TestSubsWithConfirmation(t *testing.T) {
	subs := NewWxSubs(mockUUID, mockPlan, enum.EmailLogin)

	t.Logf("Initial subscription: %+v\n", subs)

	subs, err := subs.withConfirmation(time.Now())

	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Confirmed subscripiton: %+v\n", subs)
}

func TestSubsWithMembership(t *testing.T) {
	subs := NewWxSubs(mockUUID, mockPlan, enum.EmailLogin)
	subs.IsRenewal = true

	t.Logf("Initial subscription: %+v\n", subs)

	subs, err := subs.withConfirmation(time.Now())

	if err != nil {
		t.Error(err)
		return
	}

	member := Membership{
		ExpireDate: util.DateFrom(tenDaysLater),
	}

	subs, err = subs.withMembership(member)

	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Confirmed subscription for renewal: %+v\n", subs)
}

func TestIsSubsAllowed(t *testing.T) {
	subs, err := createMember()

	subs = NewWxSubs(subs.UserID, mockPlan, enum.EmailLogin)

	ok, err := devEnv.IsSubsAllowed(subs)

	if err != nil {
		t.Error(err)
	}

	t.Logf("Is subscription allowed: %t\n", ok)
}
func TestCreateSubsNew(t *testing.T) {
	subs := NewWxSubs(mockUUID, mockPlan, enum.EmailLogin)

	err := devEnv.SaveSubscription(subs, mockClient)

	if err != nil {
		t.Error(err)
	}

	t.Logf("Saved subscription: %+v\n", subs)
}

func TestCreateSubsRenewal(t *testing.T) {
	subs, err := createMember()

	subs = NewWxSubs(subs.UserID, mockPlan, enum.EmailLogin)

	err = devEnv.SaveSubscription(subs, mockClient)

	if err != nil {
		t.Error(err)
	}

	t.Logf("Create a renewal order: %+v\n", subs)
}

func TestFindSubs(t *testing.T) {
	subs, err := createSubs()

	found, err := devEnv.FindSubscription(subs.OrderID)

	if err != nil {
		t.Error(err)
	}

	t.Logf("Found subscription: %+v\n", found)
}

func TestConfirmSubs(t *testing.T) {
	subs, err := createSubs()

	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Created subscription: %+v\n", subs)

	// Omit the find subscription step.

	subs, err = devEnv.ConfirmSubscription(subs, time.Now())

	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Confirmed subscription: %+v\n", subs)
}

func TestCreateMember(t *testing.T) {
	subs, err := createMember()

	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Subscritpion for a new member: %+v\n", subs)
}

func TestRenewMember(t *testing.T) {

	member, err := createAndFindMember()
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Created a member: %+v\n", member)

	// Create a subscription order for an existing user.
	newSubs := NewWxSubs(member.UserID, mockPlan, enum.EmailLogin)

	// Save it.
	err = devEnv.SaveSubscription(newSubs, mockClient)

	t.Logf("Save a renewal subscription")

	if err != nil {
		t.Error(err)
		return
	}

	// Confirm it. This mocks payment provider's notification.
	newSubs, err = devEnv.ConfirmSubscription(newSubs, time.Now())

	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Confirmed a renewal subscription: %+v\n", newSubs)

	err = devEnv.CreateMembership(newSubs)

	if err != nil {
		t.Error(err)
	}
}
