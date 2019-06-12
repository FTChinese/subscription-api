package controller

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gitlab.com/ftchinese/subscription-api/ali"
	"gitlab.com/ftchinese/subscription-api/wechat"
	"gitlab.com/ftchinese/subscription-api/wxlogin"
	"os"
)

const (
	wxAppMobileSubs = "wxacddf1c20516eb69" // Used by native app to pay and log in.
	wxAppMobileFTC  = "wxc1bc20ee7478536a" // Used by desktop and mobile browser to pay.
	wxAppWebFTC     = "wxc7233549ca6bc86a" // Used for web page OAuth
	wxAppFTCSupport = "wxa8e66ab05d5e212b" // Used for wechat in-house browser to pay.
)

func getWxOAuthApps() map[string]wxlogin.WxApp {
	logger := logrus.WithFields(logrus.Fields{
		"trace": "getWxOAuthApps",
	})
	var mSubs, mFTC, wFTC wxlogin.WxApp

	// 移动应用 -> FT中文网会员订阅. This is used for Android subscription
	err := viper.UnmarshalKey("wxapp.m_subs", &mSubs)
	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}
	if mSubs.Ensure() != nil {
		logger.Error("Mobile app Member subscription has empty fields")
		os.Exit(1)
	}
	// 移动应用 -> FT中文网. This is for iOS subscription and legacy Android subscription.
	err = viper.UnmarshalKey("wxapp.m_ftc", &mFTC)
	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}
	if mFTC.Ensure() != nil {
		logger.Error("Mobile app FTC has empty fields")
		os.Exit(1)
	}
	// 网站应用 -> FT中文网. This is used for web login
	err = viper.UnmarshalKey("wxapp.w_ftc", &wFTC)
	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}
	if wFTC.Ensure() != nil {
		logger.Error("Web app FTC has empty fields")
		os.Exit(1)
	}

	return map[string]wxlogin.WxApp{
		// 移动应用 -> FT中文网会员订阅. This is used for Android subscription
		wxAppMobileSubs: mSubs,
		// 移动应用 -> FT中文网. This is for iOS subscription and legacy Android subscription.
		wxAppMobileFTC: mFTC,
		// 网站应用 -> FT中文网. This is used for web login
		wxAppWebFTC: wFTC,
	}
}

func createWxpayClients() map[string]wechat.Client {
	apps := getWxPayApps()

	clients := make(map[string]wechat.Client, len(apps))

	for k, v := range apps {
		clients[k] = wechat.NewClient(v)
	}

	return clients
}

func getWxPayApps() map[string]wechat.PayApp {
	logger := logrus.WithFields(logrus.Fields{
		"package": "controller",
		"trace":   "getWxPayApps",
	})
	var mSubs, mFTC, oSupport wechat.PayApp

	// 移动应用 -> FT中文网会员订阅. This is used for Android subscription
	err := viper.UnmarshalKey("wxapp.m_subs", &mSubs)
	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}
	if mSubs.Ensure() != nil {
		logger.Error("Mobile app Member subscription has empty fields")
		os.Exit(1)
	}
	// 移动应用 -> FT中文网. This is for iOS subscription and legacy Android subscription.
	err = viper.UnmarshalKey("wxapp.m_ftc", &mFTC)
	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}
	if mFTC.Ensure() != nil {
		logger.Error("Mobile app FTC has empty fields")
		os.Exit(1)
	}

	err = viper.UnmarshalKey("wxapp.o_ftcsupport", &oSupport)
	if err != nil {
		logger.Error(err)
	}
	if oSupport.Ensure() != nil {
		logger.Error("Official account app has empty fields")
		os.Exit(1)
	}

	return map[string]wechat.PayApp{
		// 移动应用 -> FT中文网会员订阅. This is used for Android subscription
		wxAppMobileSubs: mSubs,
		// 移动应用 -> FT中文网. This is for iOS subscription and legacy Android subscription.
		wxAppMobileFTC:  mFTC,
		wxAppFTCSupport: oSupport,
	}
}

func getAliPayApp() ali.App {
	logger := logrus.WithFields(logrus.Fields{
		"package": "controller",
		"trace":   "getAliPayApp",
	})

	var app ali.App

	if err := viper.UnmarshalKey("alipay", &app); err != nil {
		logger.Error(err)
		os.Exit(1)
	}

	if err := app.Ensure(); err != nil {
		logger.Error(err)
		os.Exit(1)
	}

	return app
}
