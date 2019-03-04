package model

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/guregu/null"
	"gitlab.com/ftchinese/subscription-api/paywall"
	"testing"
	"time"
)

func TestConfirmationParcel(t *testing.T) {
	m := newMocker()

	user := m.user()
	subs := m.confirmedSubs()

	p, err := user.ConfirmationParcel(subs)
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Parcel: %+v\n", p)
}

func TestSendEmail(t *testing.T) {

	user := paywall.User{
		UserID: myFtcID,
		UserName: null.StringFrom("ToddDay"),
		Email: myFtcEmail,
	}

	subs, _ := paywall.NewWxpaySubs(
		null.StringFrom(user.UserID),
		null.String{},
		mockPlan)

	subs.ConfirmedAt = chrono.TimeNow()
	subs.IsRenewal = false
	subs.StartDate = chrono.DateNow()

	endDate, _ := subs.BillingCycle.TimeAfterACycle(time.Now())
	subs.EndDate = chrono.DateFrom(endDate)

	parcel, err := user.ConfirmationParcel(subs)

	if err != nil {
		t.Error(err)
		return
	}

	err = postman.Deliver(parcel)
	if err != nil {
		t.Error(err)
	}
}
