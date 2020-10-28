package letter

import (
	"github.com/FTChinese/go-rest/postoffice"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/subs"
)

func NewSubParcel(a reader.FtcAccount, order subs.Order) (postoffice.Parcel, error) {

	ctx := CtxSubs{
		UserName: a.NormalizeName(),
		Order:    order,
	}

	body, err := RenderNewSubs(ctx)
	if err != nil {
		return postoffice.Parcel{}, err
	}

	return postoffice.Parcel{
		FromAddress: "no-reply@ftchinese.com",
		FromName:    "FT中文网会员订阅",
		ToAddress:   a.Email,
		ToName:      ctx.UserName,
		Subject:     "会员订阅",
		Body:        body,
	}, nil
}

func NewRenewalParcel(
	a reader.FtcAccount,
	order subs.Order,
) (postoffice.Parcel, error) {

	ctx := CtxSubs{
		UserName: a.NormalizeName(),
		Order:    order,
	}

	body, err := RenderRenewalSubs(ctx)
	if err != nil {
		return postoffice.Parcel{}, err
	}

	return postoffice.Parcel{
		FromAddress: "no-reply@ftchinese.com",
		FromName:    "FT中文网会员订阅",
		ToAddress:   a.Email,
		ToName:      ctx.UserName,
		Subject:     "会员续订",
		Body:        body,
	}, nil
}

func NewUpgradeParcel(
	a reader.FtcAccount,
	order subs.Order,
	pos []subs.ProratedOrder,
) (postoffice.Parcel, error) {

	ctx := CtxUpgrade{
		UserName: a.NormalizeName(),
		Order:    order,
		Prorated: pos,
	}

	body, err := RenderUpgrade(ctx)
	if err != nil {
		return postoffice.Parcel{}, err
	}

	return postoffice.Parcel{
		FromAddress: "no-reply@ftchinese.com",
		FromName:    "FT中文网会员订阅",
		ToAddress:   a.Email,
		ToName:      ctx.UserName,
		Subject:     "会员升级",
		Body:        body,
	}, nil
}

func NewFreeUpgradeParcel(
	a reader.FtcAccount,
	order subs.Order,
	pos []subs.ProratedOrder) (postoffice.Parcel, error) {

	ctx := CtxUpgrade{
		UserName: a.NormalizeName(),
		Order:    order,
		Prorated: pos,
	}

	body, err := RenderFreeUpgrade(ctx)
	if err != nil {
		return postoffice.Parcel{}, err
	}

	return postoffice.Parcel{
		FromAddress: "no-reply@ftchinese.com",
		FromName:    "FT中文网会员订阅",
		ToAddress:   a.Email,
		ToName:      ctx.UserName,
		Subject:     "会员升级",
		Body:        body,
	}, nil
}

func NewIAPLinkParcel(account reader.FtcAccount, m reader.Membership) (postoffice.Parcel, error) {
	ctx := CtxIAPLinked{
		UserName:   account.NormalizeName(),
		Email:      account.Email,
		Tier:       m.Tier,
		ExpireDate: m.ExpireDate,
	}

	body, err := RenderIAPLinked(ctx)
	if err != nil {
		return postoffice.Parcel{}, err
	}

	return postoffice.Parcel{
		FromAddress: "no-reply@ftchinese.com",
		FromName:    "FT中文网会员订阅",
		ToAddress:   account.Email,
		ToName:      ctx.UserName,
		Subject:     "关联iOS订阅",
		Body:        body,
	}, nil
}

func NewIAPUnlinkParcel(account reader.FtcAccount, m reader.Membership) (postoffice.Parcel, error) {
	ctx := CtxIAPLinked{
		UserName:   account.NormalizeName(),
		Email:      account.Email,
		Tier:       m.Tier,
		ExpireDate: m.ExpireDate,
	}

	body, err := RenderIAPUnlinked(ctx)
	if err != nil {
		return postoffice.Parcel{}, err
	}

	return postoffice.Parcel{
		FromAddress: "no-reply@ftchinese.com",
		FromName:    "FT中文网会员订阅",
		ToAddress:   account.Email,
		ToName:      ctx.UserName,
		Subject:     "取消关联iOS订阅",
		Body:        body,
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
