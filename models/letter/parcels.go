package letter

import (
	"github.com/FTChinese/go-rest/postoffice"
	"gitlab.com/ftchinese/subscription-api/models/plan"
	"gitlab.com/ftchinese/subscription-api/models/reader"
	"gitlab.com/ftchinese/subscription-api/models/subscription"
	"strings"
	"text/template"
)

func NewSubParcel(a reader.Account, order subscription.Order) (postoffice.Parcel, error) {
	tmpl, err := template.New("order").Parse(letterNewSub)

	if err != nil {
		return postoffice.Parcel{}, err
	}

	p, err := plan.FindPlan(order.Tier, order.Cycle)
	if err != nil {
		return postoffice.Parcel{}, err
	}

	data := struct {
		User  reader.Account
		Order subscription.Order
		Plan  plan.Plan
	}{
		User:  a,
		Order: order,
		Plan:  p,
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

func NewRenewalParcel(a reader.Account, order subscription.Order) (postoffice.Parcel, error) {
	tmpl, err := template.New("order").Parse(letterRenewalSub)

	if err != nil {
		return postoffice.Parcel{}, err
	}

	p, err := plan.FindFtcPlan(order.NamedKey())
	if err != nil {
		return postoffice.Parcel{}, err
	}

	data := struct {
		User  reader.Account
		Order subscription.Order
		Plan  plan.Plan
	}{
		User:  a,
		Order: order,
		Plan:  p,
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

func NewUpgradeParcel(a reader.Account, order subscription.Order, up subscription.UpgradeSchema) (postoffice.Parcel, error) {
	tmpl, err := template.New("order").Parse(letterUpgradeSub)

	if err != nil {
		return postoffice.Parcel{}, err
	}

	p, err := plan.FindFtcPlan(order.NamedKey())
	if err != nil {
		return postoffice.Parcel{}, err
	}

	data := struct {
		User    reader.Account
		Sub     subscription.Order
		Plan    plan.Plan
		Upgrade subscription.UpgradeSchema
	}{
		User:    a,
		Sub:     order,
		Plan:    p,
		Upgrade: up,
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
//		Order  StripeSub
//		Plan Plan
//	}{
//		User: a,
//		Order:  NewStripeSub(s),
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
