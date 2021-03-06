package letter

import (
	"errors"
	"fmt"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/addon"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"strconv"
	"strings"
	"text/template"
)

var tmplCache = map[string]*template.Template{}

const (
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
	Order    subs.Order
	subs.Invoices
	Snapshot reader.MemberSnapshot
}

func (ctx CtxSubs) IsPremiumAddOn() bool {
	return ctx.Purchased.AddOnSource == addon.SourceUserPurchase && ctx.Snapshot.Tier == enum.TierPremium
}

func (ctx CtxSubs) IsSubsAddOn() bool {
	return ctx.Purchased.AddOnSource == addon.SourceUserPurchase && (ctx.Snapshot.PaymentMethod == enum.PayMethodStripe || ctx.Snapshot.PaymentMethod == enum.PayMethodApple)
}

func (ctx CtxSubs) Render() (string, error) {
	switch ctx.Order.Kind {
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

func RenderIAPLinked(ctx CtxIAPLinked) (string, error) {
	return Render(keyIAPLinked, ctx)
}

func RenderIAPUnlinked(ctx CtxIAPLinked) (string, error) {
	return Render(keyIAPUnlinked, ctx)
}
