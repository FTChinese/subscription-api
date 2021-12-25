package wxpay

import (
	"errors"
	"github.com/spf13/viper"
	"log"
)

// Source corresponds to wechat's `trade_type` fields.
type Source int

const (
	SourceDesktop Source = iota
	SourceMobile
	SourceJSAPI
	SourceApp
)

// String produces values that can be used in request to wechat.
func (x Source) String() string {
	names := [...]string{
		"NATIVE", // Pay by scanning QR
		"MWEB",   // Mobile device browser
		"JSAPI",  // Wechat embedded browser
		"APP",    // App SDK
	}

	if x < SourceDesktop || x > SourceApp {
		return ""
	}

	return names[x]
}

type AppConfig struct {
	Platform Source
	SignType string
	AppID    string `mapstructure:"app_id"`
	MchID    string `mapstructure:"mch_id"`
	APIKey   string `mapstructure:"api_key"`
}

func newAppConfig(key string) (AppConfig, error) {
	var cfg AppConfig
	err := viper.UnmarshalKey(key, &cfg)
	if err != nil {
		return AppConfig{}, err
	}

	if err := cfg.validate(); err != nil {
		return AppConfig{}, err
	}

	return cfg, nil
}

func mustNewAppConfig(key string) AppConfig {
	app, err := newAppConfig(key)
	if err != nil {
		log.Fatal(err)
	}

	return app
}

func (app AppConfig) validate() error {
	if app.AppID == "" || app.MchID == "" || app.APIKey == "" {
		return errors.New("wechat pay app_id, mch_id or secret cannot be empty")
	}

	return nil
}

func mustGetAppConfigs() []AppConfig {
	payApps := map[string]Source{
		"wxapp.app_pay":     SourceApp,
		"wxapp.jsapi_pay":   SourceJSAPI,
		"wxapp.browser_pay": SourceDesktop, // Also used as TradeTypeMobile
	}

	var apps []AppConfig
	for k, s := range payApps {
		cfg := mustNewAppConfig(k)
		cfg.Platform = s
		cfg.SignType = SignTypeMD5

		apps = append(apps, cfg)
	}

	return apps
}
