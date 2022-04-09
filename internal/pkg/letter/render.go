package letter

import (
	"errors"
	"fmt"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/internal/pkg/ftcpay"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"strconv"
	"strings"
	"text/template"
)

var tmplCache = map[string]*template.Template{}

const (
	keyVrf      = "verification"
	keyVerified = "emailVerified"
	keyWxSignUp = "wxSignUp"
	keyPwReset  = "passwordReset"
	keyLinked   = "accountLinked"
	keyUnlinkWx = "unlinkWechat"

	keyNewSubs     = "newSubs"
	keyRenewalSubs = "renewalSubs"
	keyUpgradeSubs = "upgradeSubs"
	keyAddOn       = "addOn"
	keyIAPLinked   = "iapLinked"
	keyIAPUnlinked = "iapUnlinked"
)

var funcMap = template.FuncMap{
	"formatFloat": func(f float64) string {
		return strconv.FormatFloat(f, 'f', 2, 32)
	},
	"currency": func(f float64) string {
		return fmt.Sprintf("¥ %.2f",
			f)
	},
}

func Render(name string, ctx interface{}) (string, error) {
	tmpl, ok := tmplCache[name]
	var err error
	if !ok {
		tmplStr, ok := templates[name]
		if !ok {
			return "", fmt.Errorf("template %s not found", name)
		}

		tmpl, err = template.
			New(name).
			Funcs(funcMap).
			Parse(tmplStr)

		if err != nil {
			return "", err
		}
		tmplCache[name] = tmpl
	}

	var body strings.Builder
	err = tmpl.Execute(&body, ctx)
	if err != nil {
		return "", err
	}

	return body.String(), nil
}

type CtxSubs struct {
	UserName string
	ftcpay.Invoices
}

func (ctx CtxSubs) Render() (string, error) {
	switch ctx.Purchased.OrderKind {
	case enum.OrderKindCreate:
		return Render(keyNewSubs, ctx)

	case enum.OrderKindRenew:
		return Render(keyRenewalSubs, ctx)

	case enum.OrderKindUpgrade:
		return Render(keyUpgradeSubs, ctx)

	case enum.OrderKindAddOn:
		return Render(keyAddOn, ctx)

	default:
		return "", errors.New("cannot render email for unknown subscription kind")
	}
}

type CtxIAPLinked struct {
	UserName   string
	Email      string
	Tier       enum.Tier
	ExpireDate chrono.Date
}

func (ctx CtxIAPLinked) RenderIAPLinked() (string, error) {
	return Render(keyIAPLinked, ctx)
}

func (ctx CtxIAPLinked) RenderIAPUnlinked() (string, error) {
	return Render(keyIAPUnlinked, ctx)
}

type CtxVerification struct {
	UserName string
	Email    string
	Link     string
	IsSignUp bool
}

func (ctx CtxVerification) Render() (string, error) {
	return Render(keyVrf, ctx)
}

type CtxVerified struct {
	UserName string
}

func (ctx CtxVerified) Render() (string, error) {
	return Render(keyVerified, ctx)
}

type CtxPwReset struct {
	UserName string
	URL      string
	AppCode  string
	Duration string
}

func (ctx CtxPwReset) Render() (string, error) {
	return Render(keyPwReset, ctx)
}

type CtxLinkBase struct {
	UserName   string
	Email      string
	WxNickname string
}

// CtxWxSignUp is used to render letter after a Wechat user
// create a new FTC account.
type CtxWxSignUp struct {
	CtxLinkBase
	URL string
}

func (ctx CtxWxSignUp) Render() (string, error) {
	return Render(keyWxSignUp, ctx)
}

// CtxAccountLink is used to render letter when linking/unlinking existing FTC account to wechat account.
type CtxAccountLink struct {
	CtxLinkBase
	Membership reader.Membership // The merged memebership
	FtcMember  reader.Membership // Membership prior to merging under FTC
	WxMember   reader.Membership // Membership prior to merging under wechat.
}

func (ctx CtxAccountLink) Render() (string, error) {
	return Render(keyLinked, ctx)
}

type CtxAccountUnlink struct {
	CtxLinkBase
	Membership reader.Membership
	Anchor     string
}

func (ctx CtxAccountUnlink) Render() (string, error) {
	return Render(keyUnlinkWx, ctx)
}
