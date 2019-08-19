package paywall

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"github.com/icrowley/fake"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func readJsonFile(name string) ([]byte, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	fullName := filepath.Join(dir, "../_doc", name)

	return ioutil.ReadFile(fullName)
}

func TestCWD(t *testing.T) {
	t.Log(os.Args[1])
	t.Log(os.Getwd())

	dir, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}

	fileName := filepath.Join(dir, "../_doc", "stripe_invoice.json")

	t.Log(fileName)

	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		t.Error(err)
	}

	t.Logf("%s", b)
}

func newFtcUser() Account {
	return Account{
		FtcID:    uuid.New().String(),
		UnionID:  null.String{},
		StripeID: null.String{},
		Email:    fake.EmailAddress(),
		UserName: null.StringFrom(fake.UserName()),
	}
}

func newConfirmedSub() Subscription {
	id, err := NewID(uuid.New().String(), "")
	if err != nil {
		panic(err)
	}
	s, err := NewSubs(id, standardYearlyPlan)
	if err != nil {
		panic(err)
	}

	s.PaymentMethod = enum.PayMethodAli
	s.Usage = SubsKindRenew
	s.ConfirmedAt = chrono.TimeNow()
	s.StartDate = chrono.DateNow()
	s.EndDate = chrono.DateFrom(time.Now().AddDate(1, 0, 1))

	return s
}

func TestAccount_NewSubParcel(t *testing.T) {
	parcel, err := newFtcUser().NewSubParcel(newConfirmedSub())
	if err != nil {
		t.Error(err)
	}

	t.Logf("%+v", parcel)
}

func TestAccount_RenewSubParcel(t *testing.T) {

	parcel, err := newFtcUser().RenewSubParcel(newConfirmedSub())
	if err != nil {
		t.Error(err)
	}

	t.Logf("%+v", parcel)
}

func TestAccount_UpgradeSubParcel(t *testing.T) {
	parcel, err := newFtcUser().UpgradeSubParcel(newConfirmedSub(), NewUpgradePreview(buildBalanceSources(2)))

	if err != nil {
		t.Error(err)
	}

	t.Logf("%+v", parcel)
}
