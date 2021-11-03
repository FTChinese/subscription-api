package config

import (
	"errors"
	"github.com/spf13/viper"
)

type WechatApp struct {
	AppID     string `mapstructure:"app_id"`
	AppSecret string `mapstructure:"secret"`
}

func (a WechatApp) Validate() error {
	if a.AppID == "" || a.AppSecret == "" {
		return errors.New("wechat oauth app id or secret cannot be empty")
	}

	return nil
}

func LoadWechatApp(key string) (WechatApp, error) {
	var app WechatApp
	err := viper.UnmarshalKey(key, &app)
	if err != nil {
		return WechatApp{}, err
	}

	if err := app.Validate(); err != nil {
		return WechatApp{}, err
	}

	return app, nil
}

func MustLoadWechatApp(key string) WechatApp {
	app, err := LoadWechatApp(key)
	if err != nil {
		panic(err)
	}

	return app
}

var wechatAppKeys = []string{
	// 移动应用 -> FT中文网会员订阅. This is used for Android subscription
	"wxapp.native_app",
	// 移动应用 -> FT中文网. This is for iOS subscription and legacy Android subscription.
	"wxapp.web_pay",
	// 网站应用 -> FT中文网. This is used for web login
	"wxapp.web_oauth",
}

func MustWxNativeApp() WechatApp {
	return MustLoadWechatApp(wechatAppKeys[0])
}

func MustGetWechatApps() map[string]WechatApp {
	apps := make(map[string]WechatApp)

	for _, k := range wechatAppKeys {
		app := MustLoadWechatApp(k)

		apps[app.AppID] = app
	}

	return apps
}
