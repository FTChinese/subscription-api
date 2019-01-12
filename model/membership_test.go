package model

import (
	"testing"
	"time"

	"gitlab.com/ftchinese/subscription-api/enum"
	"gitlab.com/ftchinese/subscription-api/paywall"
	"gitlab.com/ftchinese/subscription-api/util"
)

func TestCanRenew(t *testing.T) {
	member := paywall.Membership{}
	member.ExpireDate = util.DateFrom(time.Now().AddDate(1, 0, 0))

	ok := member.CanRenew(enum.CycleYear)

	t.Logf("Expire date: %s, can renew another year: %t\n", member.ExpireDate, ok)
}

func TestCannotRenew(t *testing.T) {
	member := paywall.Membership{}
	member.ExpireDate = util.DateFrom(time.Now().AddDate(1, 1, 0))

	ok := member.CanRenew(enum.CycleYear)

	t.Logf("Expire date: %s, can renew another year: %t\n", member.ExpireDate, ok)
}
func TestMemberNotFound(t *testing.T) {
	subs, _ := NewUser().CreateWxpaySubs()

	m, err := devEnv.findMember(subs)

	if err != nil {
		t.Error(err)
	}

	t.Logf("%+v\n", m)
}
