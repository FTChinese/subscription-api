package letter

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/postoffice"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/apple"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/subs"
)

var subjects = map[enum.OrderKind]string{
	enum.OrderKindCreate:  "FT会员订阅",
	enum.OrderKindRenew:   "FT会员续订",
	enum.OrderKindUpgrade: "FT会员升级",
	enum.OrderKindAddOn:   "购买FT订阅服务",
}

const fromAddress = "no-reply@ftchinese.com"

var accountKindCN = map[enum.AccountKind]string{
	enum.AccountKindFtc: "邮箱",
	enum.AccountKindWx:  "微信",
}

// VerificationParcel generates the email body for verification letter from text template.
func VerificationParcel(ctx CtxVerification) (postoffice.Parcel, error) {

	body, err := ctx.Render()
	if err != nil {
		return postoffice.Parcel{}, err
	}

	return postoffice.Parcel{
		FromAddress: fromAddress,
		FromName:    "FT中文网",
		ToName:      ctx.UserName,
		ToAddress:   ctx.Email,
		Subject:     "验证FT中文网注册邮箱",
		Body:        body,
	}, nil
}

// GreetingParcel creates a parcel to be delivered after email is verified.
func GreetingParcel(a account.BaseAccount) (postoffice.Parcel, error) {
	body, err := CtxVerified{
		UserName: a.NormalizeName(),
	}.Render()

	if err != nil {
		return postoffice.Parcel{}, err
	}

	return postoffice.Parcel{
		FromAddress: fromAddress,
		FromName:    "FT中文网",
		ToName:      a.NormalizeName(),
		ToAddress:   a.Email,
		Subject:     "邮箱验证成功",
		Body:        body,
	}, nil
}

// PasswordResetParcel generates the email body for password reset.
func PasswordResetParcel(a account.BaseAccount, session account.PwResetSession) (postoffice.Parcel, error) {

	body, err := CtxPwReset{
		UserName: a.NormalizeName(),
		URL:      session.BuildURL(),
		AppCode:  session.AppCode.String,
	}.Render()

	if err != nil {
		return postoffice.Parcel{}, err
	}

	return postoffice.Parcel{
		FromAddress: fromAddress,
		FromName:    "FT中文网",
		ToAddress:   a.Email,
		ToName:      a.NormalizeName(),
		Subject:     "[FT中文网]重置密码",
		Body:        body,
	}, nil
}

// WxSignUpParcel compose the parcel used to sent letter after wechat user creates and binds a new email account.
// Returns the parcel to be delivered by postman.
func WxSignUpParcel(a reader.Account, verifier account.EmailVerifier) (postoffice.Parcel, error) {
	body, err := CtxWxSignUp{
		CtxLinkBase: CtxLinkBase{
			UserName:   a.NormalizeName(),
			WxNickname: a.Wechat.WxNickname.String,
			Email:      a.Email,
		},
		URL: verifier.BuildURL(),
	}.Render()

	if err != nil {
		return postoffice.Parcel{}, err
	}

	return postoffice.Parcel{
		FromAddress: "no-reply@ftchinese.com",
		FromName:    "FT中文网",
		ToName:      a.NormalizeName(),
		ToAddress:   a.Email,
		Subject:     "确认FT中文网账号绑定",
		Body:        body,
	}, nil
}

// LinkedParcel generates a email parcel after accounts are linked.
func LinkedParcel(linkResult reader.LinkWxResult) (postoffice.Parcel, error) {

	body, err := CtxAccountLink{
		CtxLinkBase: CtxLinkBase{
			UserName:   linkResult.Account.NormalizeName(),
			Email:      linkResult.Account.Email,
			WxNickname: linkResult.Account.Wechat.WxNickname.String,
		},
		Membership: linkResult.Account.Membership,
		FtcMember:  linkResult.FtcMemberSnapshot.Membership,
		WxMember:   linkResult.WxMemberSnapshot.Membership,
	}.Render()

	if err != nil {
		return postoffice.Parcel{}, err
	}

	return postoffice.Parcel{
		FromAddress: "no-reply@ftchinese.com",
		FromName:    "FT中文网",
		ToName:      linkResult.Account.NormalizeName(),
		ToAddress:   linkResult.Account.Email,
		Subject:     "已绑定微信账号",
		Body:        body,
	}, nil
}

// UnlinkParcel builds an email parcel after a linked
// account severs the link.
// The receiver is the Account instance prior to unlinking.
func UnlinkParcel(a reader.Account, anchor enum.AccountKind) (postoffice.Parcel, error) {

	body, err := CtxAccountUnlink{
		CtxLinkBase: CtxLinkBase{
			UserName:   a.NormalizeName(),
			Email:      a.Email,
			WxNickname: a.Wechat.WxNickname.String,
		},
		Membership: a.Membership,
		Anchor:     accountKindCN[anchor],
	}.Render()

	if err != nil {
		return postoffice.Parcel{}, err
	}

	return postoffice.Parcel{
		FromAddress: "no-reply@ftchinese.com",
		FromName:    "FT中文网",
		ToName:      a.NormalizeName(),
		ToAddress:   a.Email,
		Subject:     "解除账号绑定",
		Body:        body,
	}, nil
}

func NewSubParcel(a account.BaseAccount, result subs.ConfirmationResult) (postoffice.Parcel, error) {

	ctx := CtxSubs{
		UserName: a.NormalizeName(),
		Order:    result.Order,
		Invoices: result.Invoices,
		Snapshot: result.Snapshot,
	}

	body, err := ctx.Render()
	if err != nil {
		return postoffice.Parcel{}, err
	}

	return postoffice.Parcel{
		FromAddress: fromAddress,
		FromName:    "FT中文网会员订阅",
		ToAddress:   a.Email,
		ToName:      ctx.UserName,
		Subject:     subjects[result.Order.Kind],
		Body:        body,
	}, nil
}

func NewIAPLinkParcel(acnt account.BaseAccount, m reader.Membership) (postoffice.Parcel, error) {
	ctx := CtxIAPLinked{
		UserName:   acnt.NormalizeName(),
		Email:      acnt.Email,
		Tier:       m.Tier,
		ExpireDate: m.ExpireDate,
	}

	body, err := RenderIAPLinked(ctx)
	if err != nil {
		return postoffice.Parcel{}, err
	}

	return postoffice.Parcel{
		FromAddress: fromAddress,
		FromName:    "FT中文网会员订阅",
		ToAddress:   acnt.Email,
		ToName:      ctx.UserName,
		Subject:     "关联iOS订阅",
		Body:        body,
	}, nil
}

func NewIAPUnlinkParcel(a account.BaseAccount, m apple.Subscription) (postoffice.Parcel, error) {
	ctx := CtxIAPLinked{
		UserName:   a.NormalizeName(),
		Email:      a.Email,
		Tier:       m.Tier,
		ExpireDate: chrono.DateFrom(m.ExpiresDateUTC.Time),
	}

	body, err := RenderIAPUnlinked(ctx)
	if err != nil {
		return postoffice.Parcel{}, err
	}

	return postoffice.Parcel{
		FromAddress: fromAddress,
		FromName:    "FT中文网会员订阅",
		ToAddress:   a.Email,
		ToName:      ctx.UserName,
		Subject:     "移除关联iOS订阅",
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
//		Price Price
//	}{
//		User: a,
//		Order:  NewStripeSub(s),
//		Price: plan,
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
//		Price    Price
//	}{
//		User:    a,
//		Invoice: i,
//		Price:    plan,
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
//		Price    Price
//	}{
//		User:    a,
//		Invoice: i,
//		Price:    plan,
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
//		Price    Price
//	}{
//		User:    a,
//		Invoice: i,
//		Price:    plan,
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
