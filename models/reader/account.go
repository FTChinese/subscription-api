package reader

import (
	"gitlab.com/ftchinese/subscription-api/models/paywall"
	"strings"
	"text/template"

	"github.com/FTChinese/go-rest/postoffice"
	"github.com/guregu/null"
)

// Account contains the minimal data to identify a user.
type Account struct {
	FtcID    string
	UnionID  null.String
	StripeID null.String
	Email    string
	UserName null.String
}

func (a Account) ID() AccountID {
	id, _ := NewID(a.FtcID, a.UnionID.String)
	return id
}

// NormalizeName returns user name, or the name part of email if name does not exist.
func (a Account) NormalizeName() string {
	if a.UserName.Valid {
		return strings.Split(a.UserName.String, "@")[0]
	}

	return strings.Split(a.Email, "@")[0]
}

// ConfirmationParcel create a parcel for email after subscription is confirmed.
func (a Account) NewSubParcel(s paywall.Subscription) (postoffice.Parcel, error) {
	tmpl, err := template.New("order").Parse(letterNewSub)

	if err != nil {
		return postoffice.Parcel{}, err
	}

	plan, err := paywall.GetFtcPlans(true).FindPlan(s.NamedKey())
	if err != nil {
		return postoffice.Parcel{}, err
	}

	data := struct {
		User Account
		Sub  paywall.Subscription
		Plan paywall.Plan
	}{
		User: a,
		Sub:  s,
		Plan: plan,
	}

	var body strings.Builder
	err = tmpl.Execute(&body, data)

	if err != nil {
		return postoffice.Parcel{}, err
	}

	return postoffice.Parcel{
		FromAddress: "no-reply@ftchinese.com",
		FromName:    "FT中文网会员订阅",
		ToAddress:   a.Email,
		ToName:      a.NormalizeName(),
		Subject:     "会员订阅",
		Body:        body.String(),
	}, nil
}

func (a Account) RenewSubParcel(s paywall.Subscription) (postoffice.Parcel, error) {
	tmpl, err := template.New("order").Parse(letterRenewalSub)

	if err != nil {
		return postoffice.Parcel{}, err
	}

	plan, err := paywall.GetFtcPlans(true).FindPlan(s.NamedKey())
	if err != nil {
		return postoffice.Parcel{}, err
	}

	data := struct {
		User Account
		Sub  paywall.Subscription
		Plan paywall.Plan
	}{
		User: a,
		Sub:  s,
		Plan: plan,
	}

	var body strings.Builder
	err = tmpl.Execute(&body, data)

	if err != nil {
		return postoffice.Parcel{}, err
	}

	return postoffice.Parcel{
		FromAddress: "no-reply@ftchinese.com",
		FromName:    "FT中文网会员订阅",
		ToAddress:   a.Email,
		ToName:      a.NormalizeName(),
		Subject:     "会员续订",
		Body:        body.String(),
	}, nil
}

func (a Account) UpgradeSubParcel(s paywall.Subscription, preview paywall.UpgradePlan) (postoffice.Parcel, error) {
	tmpl, err := template.New("order").Parse(letterUpgradeSub)

	if err != nil {
		return postoffice.Parcel{}, err
	}

	plan, err := paywall.GetFtcPlans(true).FindPlan(s.NamedKey())
	if err != nil {
		return postoffice.Parcel{}, err
	}

	data := struct {
		User    Account
		Sub     paywall.Subscription
		Plan    paywall.Plan
		Upgrade paywall.UpgradePlan
	}{
		User:    a,
		Sub:     s,
		Plan:    plan,
		Upgrade: preview,
	}

	var body strings.Builder
	err = tmpl.Execute(&body, data)

	if err != nil {
		return postoffice.Parcel{}, err
	}

	return postoffice.Parcel{
		FromAddress: "no-reply@ftchinese.com",
		FromName:    "FT中文网会员订阅",
		ToAddress:   a.Email,
		ToName:      a.NormalizeName(),
		Subject:     "会员升级",
		Body:        body.String(),
	}, nil
}

//func (a Account) StripeSubParcel(s *stripe.Subscription) (postoffice.Parcel, error) {
//	tmpl, err := template.New("stripe_sub").Parse(letterStripeSub)
//
//	if err != nil {
//		return postoffice.Parcel{}, err
//	}
//
//	plan, _ := BuildFtcPlanForStripe(s)
//	data := struct {
//		User Account
//		Sub  StripeSub
//		Plan Plan
//	}{
//		User: a,
//		Sub:  NewStripeSub(s),
//		Plan: plan,
//	}
//
//	var body strings.Builder
//	err = tmpl.Execute(&body, data)
//
//	if err != nil {
//		return postoffice.Parcel{}, err
//	}
//
//	return postoffice.Parcel{
//		FromAddress: "no-reply@ftchinese.com",
//		FromName:    "FT中文网会员订阅",
//		ToAddress:   a.Email,
//		ToName:      a.NormalizeName(),
//		Subject:     "Stripe订阅",
//		Body:        body.String(),
//	}, nil
//}

//func (a Account) StripeInvoiceParcel(i StripeInvoice) (postoffice.Parcel, error) {
//	tmpl, err := template.New("stripe_invoice").Parse(letterStripeInvoice)
//
//	if err != nil {
//		return postoffice.Parcel{}, err
//	}
//
//	plan, _ := i.BuildFtcPlan()
//	data := struct {
//		User    Account
//		Invoice StripeInvoice
//		Plan    Plan
//	}{
//		User:    a,
//		Invoice: i,
//		Plan:    plan,
//	}
//
//	var body strings.Builder
//	err = tmpl.Execute(&body, data)
//
//	if err != nil {
//		return postoffice.Parcel{}, err
//	}
//
//	return postoffice.Parcel{
//		FromAddress: "no-reply@ftchinese.com",
//		FromName:    "FT中文网会员订阅",
//		ToAddress:   a.Email,
//		ToName:      a.NormalizeName(),
//		Subject:     "Stripe订阅发票",
//		Body:        body.String(),
//	}, nil
//}
//
//func (a Account) StripePaymentFailed(i StripeInvoice) (postoffice.Parcel, error) {
//	tmpl, err := template.New("stripe_payment_failed").Parse(letterStripePaymentFailed)
//
//	if err != nil {
//		return postoffice.Parcel{}, err
//	}
//
//	plan, _ := i.BuildFtcPlan()
//	data := struct {
//		User    Account
//		Invoice StripeInvoice
//		Plan    Plan
//	}{
//		User:    a,
//		Invoice: i,
//		Plan:    plan,
//	}
//
//	var body strings.Builder
//	err = tmpl.Execute(&body, data)
//
//	if err != nil {
//		return postoffice.Parcel{}, err
//	}
//
//	return postoffice.Parcel{
//		FromAddress: "no-reply@ftchinese.com",
//		FromName:    "FT中文网会员订阅",
//		ToAddress:   a.Email,
//		ToName:      a.NormalizeName(),
//		Subject:     "Stripe支付失败",
//		Body:        body.String(),
//	}, nil
//}
//
//func (a Account) StripeActionRequired(i StripeInvoice) (postoffice.Parcel, error) {
//	tmpl, err := template.New("stripe_action_required").Parse(letterPaymentActionRequired)
//
//	if err != nil {
//		return postoffice.Parcel{}, err
//	}
//
//	plan, _ := i.BuildFtcPlan()
//	data := struct {
//		User    Account
//		Invoice StripeInvoice
//		Plan    Plan
//	}{
//		User:    a,
//		Invoice: i,
//		Plan:    plan,
//	}
//
//	var body strings.Builder
//	err = tmpl.Execute(&body, data)
//
//	if err != nil {
//		return postoffice.Parcel{}, err
//	}
//
//	return postoffice.Parcel{
//		FromAddress: "no-reply@ftchinese.com",
//		FromName:    "FT中文网会员订阅",
//		ToAddress:   a.Email,
//		ToName:      a.NormalizeName(),
//		Subject:     "Stripe支付尚未完成",
//		Body:        body.String(),
//	}, nil
//}
