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
	subs, err := createMember(false)

	subs = NewWxSubs(subs.UserID, mockPlan, enum.EmailLogin)

	ok, err := devEnv.IsSubsAllowed(subs)

	if err != nil {
		t.Error(err)
	}

	t.Logf("Is subscription allowed: %t\n", ok)
}
func TestCreateSubs(t *testing.T) {
	subs := NewWxSubs(mockUUID, mockPlan, enum.EmailLogin)

	err := devEnv.SaveSubscription(subs, mockClient)

	if err != nil {
		t.Error(err)
	}

	t.Logf("Saved subscription: %+v\n", subs)
}

func TestFindSubs(t *testing.T) {
	subs, err := createSubs(false)

	found, err := devEnv.FindSubscription(subs.OrderID)

	if err != nil {
		t.Error(err)
	}

	t.Logf("Found subscription: %+v\n", found)
}

func TestConfirmSubs(t *testing.T) {
	subs, err := createSubs(false)

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
	subs, err := createMember(false)

	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Subscritpion for a new member: %+v\n", subs)
}

func TestRenewMember(t *testing.T) {
	// First iteration creates a new subscription,
	// second iteration renew the membership.
	for i := 0; i < 2; i++ {
		subs, err := createMember(false)
		if err != nil {
			t.Error(err)
			break
		}

		t.Logf("Subscripiton for a membership: %+v\n", subs)
	}
}

func TestCreatemember_wxLogin(t *testing.T) {
	for i := 0; i < 2; i++ {
		subs, err := createMember(true)
		if err != nil {
			t.Error(err)
			break
		}

		t.Logf("Subscription for a membership: %+v\n", subs)
	}
}
