package letter

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/internal/pkg/apple"
	"github.com/FTChinese/subscription-api/internal/pkg/subs"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/pkg/postman"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"go.uber.org/zap"
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

type Service struct {
	postman postman.Postman
	logger  *zap.Logger
}

func NewService(logger *zap.Logger) Service {
	return Service{
		postman: postman.New(config.MustGetHanqiConn()),
		logger:  logger,
	}
}

// SendVerification generates the email body for verification letter from text template.
func (s Service) SendVerification(ctx CtxVerification) error {
	defer s.logger.Sync()
	sugar := s.logger.Sugar()

	body, err := ctx.Render()
	if err != nil {
		sugar.Error(err)
		return err
	}

	parcel := postman.Parcel{
		FromAddress: fromAddress,
		FromName:    "FT中文网",
		ToName:      ctx.UserName,
		ToAddress:   ctx.Email,
		Subject:     "验证FT中文网注册邮箱",
		Body:        body,
	}

	sugar.Info(parcel)

	return s.postman.Deliver(parcel)
}

// SendGreeting creates a parcel to be delivered after email is verified.
func (s Service) SendGreeting(a account.BaseAccount) error {
	defer s.logger.Sync()
	sugar := s.logger.Sugar()

	body, err := CtxVerified{
		UserName: a.NormalizeName(),
	}.Render()

	if err != nil {
		sugar.Error(err)
		return err
	}

	parcel := postman.Parcel{
		FromAddress: fromAddress,
		FromName:    "FT中文网",
		ToName:      a.NormalizeName(),
		ToAddress:   a.Email,
		Subject:     "邮箱验证成功",
		Body:        body,
	}

	sugar.Info(parcel)

	return s.postman.Deliver(parcel)
}

// SendPasswordReset generates the email body for password reset.
func (s Service) SendPasswordReset(a account.BaseAccount, session account.PwResetSession) error {
	defer s.logger.Sync()
	sugar := s.logger.Sugar()

	body, err := CtxPwReset{
		UserName: a.NormalizeName(),
		URL:      session.BuildURL(),
		AppCode:  session.AppCode.String,
		Duration: session.FormatDuration(),
	}.Render()

	if err != nil {
		sugar.Error(err)
		return err
	}

	parcel := postman.Parcel{
		FromAddress: fromAddress,
		FromName:    "FT中文网",
		ToAddress:   a.Email,
		ToName:      a.NormalizeName(),
		Subject:     "[FT中文网]重置密码",
		Body:        body,
	}

	sugar.Info(parcel)

	return s.postman.Deliver(parcel)
}

// SendWxSignUp sends an email after wechat-user linked to a new
// email account.
func (s Service) SendWxSignUp(a reader.Account, v account.EmailVerifier) error {
	defer s.logger.Sync()
	sugar := s.logger.Sugar()

	body, err := CtxWxSignUp{
		CtxLinkBase: CtxLinkBase{
			UserName:   a.NormalizeName(),
			WxNickname: a.Wechat.WxNickname.String,
			Email:      a.Email,
		},
		URL: v.BuildURL(),
	}.Render()

	if err != nil {
		sugar.Error()
		return err
	}

	parcel := postman.Parcel{
		FromAddress: "no-reply@ftchinese.com",
		FromName:    "FT中文网",
		ToName:      a.NormalizeName(),
		ToAddress:   a.Email,
		Subject:     "确认FT中文网账号绑定",
		Body:        body,
	}

	sugar.Info(parcel)

	return s.postman.Deliver(parcel)
}

// SendWxEmailLink sends an email to user after
// either email account linking wechat,
// or wechat linking existing email
func (s Service) SendWxEmailLink(linkResult reader.WxEmailLinkResult) error {
	defer s.logger.Sync()
	sugar := s.logger.Sugar()

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
		sugar.Error(err)
		return err
	}

	parcel := postman.Parcel{
		FromAddress: "no-reply@ftchinese.com",
		FromName:    "FT中文网",
		ToName:      linkResult.Account.NormalizeName(),
		ToAddress:   linkResult.Account.Email,
		Subject:     "已绑定微信账号",
		Body:        body,
	}

	sugar.Info(parcel)

	return s.postman.Deliver(parcel)
}

// SendWxEmailUnlink builds an email parcel after a linked
// account severs the link.
func (s Service) SendWxEmailUnlink(a reader.Account, anchor enum.AccountKind) error {
	defer s.logger.Sync()
	sugar := s.logger.Sugar()

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
		sugar.Error(err)
		return err
	}

	parcel := postman.Parcel{
		FromAddress: "no-reply@ftchinese.com",
		FromName:    "FT中文网",
		ToName:      a.NormalizeName(),
		ToAddress:   a.Email,
		Subject:     "解除账号绑定",
		Body:        body,
	}

	return s.postman.Deliver(parcel)
}

// SendOneTimePurchase sends an email after user made a
// successful one-time purchase.
func (s Service) SendOneTimePurchase(a account.BaseAccount, invs subs.Invoices) error {
	defer s.logger.Sync()
	sugar := s.logger.Sugar()

	body, err := CtxSubs{
		UserName: a.NormalizeName(),
		Invoices: invs,
	}.Render()

	if err != nil {
		sugar.Error(err)
		return err
	}

	parcel := postman.Parcel{
		FromAddress: fromAddress,
		FromName:    "FT中文网会员订阅",
		ToAddress:   a.Email,
		ToName:      a.NormalizeName(),
		Subject:     subjects[invs.Purchased.OrderKind],
		Body:        body,
	}

	return s.postman.Deliver(parcel)
}

func (s Service) SendIAPLinked(a account.BaseAccount, m reader.Membership) error {
	defer s.logger.Sync()
	sugar := s.logger.Sugar()

	body, err := CtxIAPLinked{
		UserName:   a.NormalizeName(),
		Email:      a.Email,
		Tier:       m.Tier,
		ExpireDate: m.ExpireDate,
	}.RenderIAPLinked()

	if err != nil {
		sugar.Error(err)
		return err
	}

	parcel := postman.Parcel{
		FromAddress: fromAddress,
		FromName:    "FT中文网会员订阅",
		ToAddress:   a.Email,
		ToName:      a.NormalizeName(),
		Subject:     "关联iOS订阅",
		Body:        body,
	}

	return s.postman.Deliver(parcel)
}

func (s Service) SendIAPUnlinked(a account.BaseAccount, m apple.Subscription) error {
	defer s.logger.Sync()
	sugar := s.logger.Sugar()

	body, err := CtxIAPLinked{
		UserName:   a.NormalizeName(),
		Email:      a.Email,
		Tier:       m.Tier,
		ExpireDate: chrono.DateFrom(m.ExpiresDateUTC.Time),
	}.RenderIAPUnlinked()

	if err != nil {
		sugar.Error(err)
		return err
	}

	parcel := postman.Parcel{
		FromAddress: fromAddress,
		FromName:    "FT中文网会员订阅",
		ToAddress:   a.Email,
		ToName:      a.NormalizeName(),
		Subject:     "移除关联iOS订阅",
		Body:        body,
	}

	return s.postman.Deliver(parcel)
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
