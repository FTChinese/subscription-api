package letter

import (
	"fmt"
	"strconv"
	"strings"
	"text/template"
)

var tmplCache = map[string]*template.Template{}

const (
	keyNewSubs     = "newSubs"
	keyRenewalSubs = "renewalSubs"
	keyUpgradeSubs = "upgradeSubs"
	keyFreeUpgrade = "freeUpgrade"
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

func RenderNewSubs(ctx CtxSubs) (string, error) {
	return Render(keyNewSubs, ctx)
}

func RenderRenewalSubs(ctx CtxSubs) (string, error) {
	return Render(keyRenewalSubs, ctx)
}

func RenderUpgrade(ctx CtxUpgrade) (string, error) {
	return Render(keyUpgradeSubs, ctx)
}

func RenderFreeUpgrade(ctx CtxUpgrade) (string, error) {
	return Render(keyFreeUpgrade, ctx)
}

func RenderIAPLinked(ctx CtxIAPLinked) (string, error) {
	return Render(keyIAPLinked, ctx)
}

func RenderIAPUnlinked(ctx CtxIAPLinked) (string, error) {
	return Render(keyIAPUnlinked, ctx)
}
