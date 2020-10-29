package wechat

import (
	"errors"
	"github.com/spf13/viper"
	"log"
)

type appConfig struct {
	Platform TradeType
	Key      string
}

var appCfgs = []appConfig{
	{
		Platform: TradeTypeApp,
		Key:      "wxapp.native_app",
	},
	{
		Platform: TradeTypeJSAPI,
		Key:      "wxapp.webrowser_pay",
	},
	{
		Platform: TradeTypeDesktop,
		Key:      "wxapp.web_pay",
	},
}

type PayApp struct {
	Platform TradeType
	AppID    string `mapstructure:"app_id"`
	MchID    string `mapstructure:"mch_id"`
	APIKey   string `mapstructure:"api_key"`
}

func NewPayApp(key string) (PayApp, error) {
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
		log.Fatal(err)
	}

	return app
}

func (a PayApp) Validate() error {
	if a.AppID == "" || a.MchID == "" || a.APIKey == "" {
		return errors.New("wechat pay app_id, mch_id or secret cannot be empty")
	}

	return nil
}

func MustGetPayApps() []PayApp {
	keys := map[string]TradeType{
		"wxapp.native_app":    TradeTypeApp,
		"wxapp.webrowser_pay": TradeTypeJSAPI,
		"wxapp.web_pay":       TradeTypeDesktop, // Also used as TradeTypeMobile
	}

	var apps []PayApp
	for k, p := range keys {
		app := MustNewPayApp(k)
		app.Platform = p
		apps = append(apps, app)
	}

	return apps
}
