package controller

import (
	"github.com/spf13/viper"
	"gitlab.com/ftchinese/subscription-api/ali"
	"gitlab.com/ftchinese/subscription-api/wechat"
	"gitlab.com/ftchinese/subscription-api/wxlogin"
	"os"
)

const (
	wxAppMobileSubs = "***REMOVED***"
	wxAppMobileFTC  = "***REMOVED***"
	wxAppWebFTC     = "wxc7233549ca6bc86a"
)

func getWxOAuthApps() map[string]wxlogin.WxApp {
	var mSubs, mFTC, wFTC wxlogin.WxApp

	// 移动应用 -> FT中文网会员订阅. This is used for Android subscription
	err := viper.UnmarshalKey("wxapp.m_subs", &mSubs)
	if err != nil {
		logger.WithField("trace", "wxOAuthApps").Error(err)
		os.Exit(1)
	}
	if mSubs.Ensure() != nil {
		logger.WithField("trace", "wxOAuthApps").Error("Mobile app Member subscription has empty fields")
		os.Exit(1)
	}
	// 移动应用 -> FT中文网. This is for iOS subscription and legacy Android subscription.
	err = viper.UnmarshalKey("wxapp.m_ftc", &mFTC)
	if err != nil {
		logger.WithField("trace", "wxOAuthApps").Error(err)
		os.Exit(1)
	}
	if mFTC.Ensure() != nil {
		logger.WithField("trace", "wxOAuthApps").Error("Mobile app FTC has empty fields")
		os.Exit(1)
	}
	// 网站应用 -> FT中文网. This is used for web login
	err = viper.UnmarshalKey("wxapp.w_ftc", &wFTC)
	if err != nil {
		logger.WithField("trace", "wxOAuthApps").Error(err)
		os.Exit(1)
	}
	if wFTC.Ensure() != nil {
		logger.WithField("trace", "wxOAuthApps").Error("Web app FTC has empty fields")
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

func getWxPayApps() map[string]wechat.PayApp {
	var mSubs, mFTC wechat.PayApp

	// 移动应用 -> FT中文网会员订阅. This is used for Android subscription
	err := viper.UnmarshalKey("wxapp.m_subs", &mSubs)
	if err != nil {
		logger.WithField("trace", "wxOAuthApps").Error(err)
		os.Exit(1)
	}
	if mSubs.Ensure() != nil {
		logger.WithField("trace", "wxOAuthApps").Error("Mobile app Member subscription has empty fields")
		os.Exit(1)
	}
	// 移动应用 -> FT中文网. This is for iOS subscription and legacy Android subscription.
	err = viper.UnmarshalKey("wxapp.m_ftc", &mFTC)
	if err != nil {
		logger.WithField("trace", "wxOAuthApps").Error(err)
		os.Exit(1)
	}
	if mFTC.Ensure() != nil {
		logger.WithField("trace", "wxOAuthApps").Error("Mobile app FTC has empty fields")
		os.Exit(1)
	}

	return map[string]wechat.PayApp{
		// 移动应用 -> FT中文网会员订阅. This is used for Android subscription
		wxAppMobileSubs: mSubs,
		// 移动应用 -> FT中文网. This is for iOS subscription and legacy Android subscription.
		wxAppMobileFTC: mFTC,
	}
}

func getAliPayApp() ali.App {
	var app ali.App

	if err := viper.UnmarshalKey("alipay", &app); err != nil {
		logger.WithField("trace", "NewAliRouter").Error(err)
		os.Exit(1)
	}

	if err := app.Ensure(); err != nil {
		logger.WithField("trace", "NewAliRouter").Error(err)
		os.Exit(1)
	}

	return app
}
