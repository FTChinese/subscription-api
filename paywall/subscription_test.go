package paywall

import (
	"github.com/FTChinese/go-rest"
	"github.com/guregu/null"
	"github.com/satori/go.uuid"
	"testing"
	"time"
)

func genWxID() string {
	id, _ := gorest.RandomBase64(21)
	return id
}

func genUserID() string {
	return uuid.Must(uuid.NewV4()).String()
}

var plan = defaultPlans["standard_year"]

func TestNewSubs(t *testing.T) {
	ftcID := null.StringFrom(genUserID())
	unionID := null.StringFrom(genWxID())

	subs, err := NewWxpaySubs(ftcID, unionID, plan)
	if err != nil {
		t.Error(err)
		return
	}
	if subs.CompoundID != ftcID.String {
		t.Errorf("Expected user id %s, got %s", ftcID.String, subs.CompoundID)
	}

	t.Logf("Wxpay order: %+v\n", subs)

	subs, err = NewAlipaySubs(ftcID, unionID, plan)
	if err != nil {
		t.Error(err)
		return
	}
	if subs.CompoundID != ftcID.String {
		t.Errorf("Expected user id %s, got %s", ftcID.String, subs.CompoundID)
	}
	t.Logf("Alipay order: %+v\n", subs)
}

func TestSubscription_Methods(t *testing.T) {
	subs := Subscription{}
	err := subs.generateOrderID()
	if err != nil {
		t.Error(err)
		return
	}
	if subs.OrderID == "" {
		t.Error("Order ID should not be empty")
		return
	}
	t.Logf("Order ID: %s\n", subs.OrderID)

	subs, _ = NewWxpaySubs(
		null.StringFrom(genUserID()),
		null.StringFrom(genWxID()),
		plan)

	t.Logf("Ali net price: %s\n", subs.AliNetPrice())
	t.Logf("Wx net price: %d\n", subs.WxNetPrice())

	if subs.IsConfirmed() {
		t.Errorf("Order should be confirmed: %t\n", subs.IsConfirmed())
		return
	}

	subs, err = subs.ConfirmWithDuration(Duration{}, time.Now())
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Confirmed order: %+v\n", subs)
}
