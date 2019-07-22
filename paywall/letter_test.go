package paywall

import (
	"encoding/json"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go"
	"strings"
	"testing"
	"text/template"
	"time"
)

func TestLetter_NewSub(t *testing.T) {
	tmpl, err := template.New("confirmation").Parse(letterNewSub)
	if err != nil {
		t.Error(err)
	}

	id, _ := NewID(uuid.New().String(), "")
	s, _ := NewSubs(id, standardYearlyPlan)
	s.PaymentMethod = enum.PayMethodAli
	s.Usage = SubsKindCreate
	s.ConfirmedAt = chrono.TimeNow()
	s.StartDate = chrono.DateNow()
	s.EndDate = chrono.DateFrom(time.Now().AddDate(1, 0, 1))

	var body strings.Builder
	data := struct {
		User FtcUser
		Sub  Subscription
		Plan Plan
	}{
		User: newFtcUser(),
		Sub:  s,
		Plan: standardYearlyPlan,
	}
	err = tmpl.Execute(&body, data)

	if err != nil {
		t.Error(err)
	}

	t.Log(body.String())
}

func TestLetter_RenewSub(t *testing.T) {
	tmpl, err := template.New("confirmation").Parse(letterRenewalSub)
	if err != nil {
		t.Error(err)
	}

	id, _ := NewID(uuid.New().String(), "")
	s, _ := NewSubs(id, standardYearlyPlan)
	s.PaymentMethod = enum.PayMethodAli
	s.Usage = SubsKindRenew
	s.ConfirmedAt = chrono.TimeNow()
	s.StartDate = chrono.DateNow()
	s.EndDate = chrono.DateFrom(time.Now().AddDate(1, 0, 1))

	var body strings.Builder
	data := struct {
		User FtcUser
		Sub  Subscription
		Plan Plan
	}{
		User: newFtcUser(),
		Sub:  s,
		Plan: standardYearlyPlan,
	}
	err = tmpl.Execute(&body, data)

	if err != nil {
		t.Error(err)
	}

	t.Log(body.String())
}

func TestLetter_UpgradeSub(t *testing.T) {
	tmpl, err := template.New("confirmation").Parse(letterUpgradeSub)
	if err != nil {
		t.Error(err)
	}

	id, _ := NewID(uuid.New().String(), "")
	s, _ := NewSubs(id, premiumYearlyPlan)
	s.PaymentMethod = enum.PayMethodAli
	s.Usage = SubsKindUpgrade
	s.ConfirmedAt = chrono.TimeNow()
	s.StartDate = chrono.DateNow()
	s.EndDate = chrono.DateFrom(time.Now().AddDate(1, 0, 1))

	var body strings.Builder
	data := struct {
		User    FtcUser
		Sub     Subscription
		Plan    Plan
		Upgrade UpgradePreview
	}{
		User:    newFtcUser(),
		Sub:     s,
		Plan:    standardYearlyPlan,
		Upgrade: NewUpgradePreview(buildBalanceSources(2)),
	}

	err = tmpl.Execute(&body, data)

	if err != nil {
		t.Error(err)
	}

	t.Log(body.String())
}

func TestLetter_StripeSubscription(t *testing.T) {
	tmpl, err := template.New("stripe_sub").Parse(letterStripeSub)
	if err != nil {
		t.Error(err)
	}

	s := stripe.Subscription{}

	if err := json.Unmarshal([]byte(subData), &s); err != nil {
		t.Error(err)
	}

	ss := NewStripeSub(&s)

	plan, err := ss.BuildFtcPlan()
	if err != nil {
		t.Error(err)
	}
	var body strings.Builder
	data := struct {
		User FtcUser
		Sub  StripeSub
		Plan Plan
	}{
		User: newFtcUser(),
		Sub:  ss,
		Plan: plan,
	}
	err = tmpl.Execute(&body, data)

	if err != nil {
		t.Error(err)
	}

	t.Log(body.String())
}

func TestLetter_StripeInvoice(t *testing.T) {
	var i stripe.Invoice

	if err := json.Unmarshal([]byte(invoiceData), &i); err != nil {
		t.Error(err)
	}

	tmpl, err := template.New("invoice").Parse(letterStripeInvoice)
	if err != nil {
		t.Error(err)
	}

	ei := EmailedInvoice{&i}

	plan, err := ei.BuildFtcPlan()
	if err != nil {
		t.Error(err)
	}

	var body strings.Builder
	data := struct {
		User    FtcUser
		Invoice EmailedInvoice
		Plan    Plan
	}{
		User:    newFtcUser(),
		Invoice: ei,
		Plan:    plan,
	}
	err = tmpl.Execute(&body, data)

	if err != nil {
		t.Error(err)
	}

	t.Log(body.String())
}
