package paywall

import (
	"github.com/pkg/errors"
	"strings"
	"text/template"

	"github.com/FTChinese/go-rest/postoffice"
	"github.com/guregu/null"
)

// FtcID is used to identify an FTC user.
// A user might have an ftc uuid, or a wechat union id,
// or both.
// This type structure is used to ensure unique constraint
// for SQL columns that cannot be both null since SQL do not
// have a mechanism to do UNIQUE INDEX on two columns while
// keeping either of them nullable.
// A user's compound id is taken from either ftc uuid or
// wechat id, with ftc id taking precedence.
type AccountID struct {
	CompoundID string      `json:"-"`
	FtcID      null.String `json:"-"`
	UnionID    null.String `json:"-"`
}

func NewID(ftcID, unionID string) (AccountID, error) {
	id := AccountID{
		FtcID:   null.NewString(ftcID, ftcID != ""),
		UnionID: null.NewString(unionID, unionID != ""),
	}

	if ftcID != "" {
		id.CompoundID = ftcID
	} else if unionID != "" {
		id.CompoundID = unionID
	} else {
		return id, errors.New("ftcID and unionID should not both be null")
	}
	return id, nil
}

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
func (a Account) NewSubParcel(s Subscription) (postoffice.Parcel, error) {
	tmpl, err := template.New("order").Parse(letterNewSub)

	if err != nil {
		return postoffice.Parcel{}, err
	}

	plan, err := GetFtcPlans(true).FindPlan(s.NamedKey())
	if err != nil {
		return postoffice.Parcel{}, err
	}

	data := struct {
		User Account
		Sub  Subscription
		Plan Plan
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

func (a Account) RenewSubParcel(s Subscription) (postoffice.Parcel, error) {
	tmpl, err := template.New("order").Parse(letterRenewalSub)

	if err != nil {
		return postoffice.Parcel{}, err
	}

	plan, err := GetFtcPlans(true).FindPlan(s.NamedKey())
	if err != nil {
		return postoffice.Parcel{}, err
	}

	data := struct {
		User Account
		Sub  Subscription
		Plan Plan
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

func (a Account) UpgradeSubParcel(s Subscription, preview UpgradePreview) (postoffice.Parcel, error) {
	tmpl, err := template.New("order").Parse(letterUpgradeSub)

	if err != nil {
		return postoffice.Parcel{}, err
	}

	plan, err := GetFtcPlans(true).FindPlan(s.NamedKey())
	if err != nil {
		return postoffice.Parcel{}, err
	}

	data := struct {
		User    Account
		Sub     Subscription
		Plan    Plan
		Upgrade UpgradePreview
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
