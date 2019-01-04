package model

import (
	"testing"
	"time"

	"gitlab.com/ftchinese/subscription-api/enum"
	"gitlab.com/ftchinese/subscription-api/util"
)

func TestCanRenew(t *testing.T) {
	member := Membership{
		ExpireDate: util.DateFrom(time.Now().AddDate(1, 0, 0)),
	}

	ok := member.CanRenew(enum.CycleYear)

	t.Logf("Expire date: %s, can renew another year: %t\n", member.ExpireDate, ok)
}

func TestCannotRenew(t *testing.T) {
	member := Membership{
		ExpireDate: util.DateFrom(time.Now().AddDate(1, 1, 0)),
	}

	ok := member.CanRenew(enum.CycleYear)

	t.Logf("Expire date: %s, can renew another year: %t\n", member.ExpireDate, ok)
}
func TestMemberNotFound(t *testing.T) {
	subs := NewWxSubs(mockUUID, mockPlan, enum.EmailLogin)

	m, err := devEnv.findMember(subs)

	if err != nil {
		t.Error(err)
	}

	t.Logf("%+v\n", m)
}

func TestFoundMember(t *testing.T) {
	subs := NewWxSubs(mockUUID, mockPlan, enum.EmailLogin)

	subs, err := subs.withConfirmation(time.Now())

	err = devEnv.CreateMembership(subs)

	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Created a member from subscripiton: %+v\n", subs)

	m, err := devEnv.findMember(subs)

	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Find membership: %+v\n", m)
}
