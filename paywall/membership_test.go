package paywall

import (
	"testing"
	"time"

	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
)

func TestCanRenew(t *testing.T) {
	member := Membership{}

	member.ExpireDate = chrono.DateFrom(time.Now().AddDate(1, 0, 0))

	ok := member.CanRenew(enum.CycleYear)

	t.Logf("Expire date: %s, can renew another year: %t\n", member.ExpireDate, ok)
}

func TestCannotRenew(t *testing.T) {
	member := Membership{}
	member.ExpireDate = chrono.DateFrom(time.Now().AddDate(1, 1, 0))

	ok := member.CanRenew(enum.CycleYear)

	t.Logf("Expire date: %s, can renew another year: %t\n", member.ExpireDate, ok)
}
