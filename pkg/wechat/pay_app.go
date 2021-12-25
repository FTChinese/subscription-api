package wechat

import (
	"errors"
	"fmt"
	"github.com/spf13/viper"
)

type PayApp struct {
	Platform TradeType
	AppID    string `mapstructure:"app_id"`
	MchID    string `mapstructure:"mch_id"`
	APIKey   string `mapstructure:"api_key"`
}

func NewPayApp(key string) (PayApp, error) {
	fmt.Printf("Initializing wx app %s\n", key)
	var app PayApp
	err := viper.UnmarshalKey(key, &app)
	if err != nil {
		return PayApp{}, err
	}

	if err := app.Validate(); err != nil {
		return PayApp{}, err
	}

	return app, nil
}

func MustNewPayApp(key string) PayApp {
	app, err := NewPayApp(key)
	if err != nil {
		panic(err)
	}

	return app
}

func (app PayApp) Validate() error {
	if app.AppID == "" || app.MchID == "" || app.APIKey == "" {
		return errors.New("wechat pay app_id, mch_id or secret cannot be empty")
	}

	return nil
}

func MustGetPayApps() []PayApp {
	keys := map[string]TradeType{
		"wxapp.app_pay":     TradeTypeApp,
		"wxapp.jsapi_pay":   TradeTypeJSAPI,
		"wxapp.browser_pay": TradeTypeDesktop, // Also used as TradeTypeMobile
	}

	var apps []PayApp
	for k, p := range keys {
		app := MustNewPayApp(k)
		app.Platform = p
		apps = append(apps, app)
	}

	return apps
}
