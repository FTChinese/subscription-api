package config

import "github.com/FTChinese/go-rest/enum"

const readerAppBase = "https://next.ftacademy.cn"

const (
	// EmailVerificationURL is the base url to construct url in email verification letter.
	// Previously we used https://users.chineseft.com/verify/email created by the next-user app, which is no longer maintained.
	EmailVerificationURL = readerAppBase + "/reader/verification"
	// PasswordResetURL is the base url to construct url to reset password.
	// Previously we used https://users.ftchinese.com/password-reset created by the next-user app.
	PasswordResetURL = readerAppBase + "/reader/password-reset"
)

// AliWxWebhookURL builds the url for one-time purchase.
// If isProd is true, use online production server;
// otherwise always use sandbox server.
// The webhook url for one-off purchase:
// https://www.ftacademy.cn/api/<v6|sandbox>/webhook/<wxpay|alipay>
func AliWxWebhookURL(isProd bool, method enum.PayMethod) string {
	var baseURL = "https://www.ftacademy.cn"
	var v string
	var m string
	if isProd {
		v = Version
	} else {
		v = "sandbox"
	}

	switch method {
	case enum.PayMethodWx:
		m = "wxpay"
	case enum.PayMethodAli:
		m = "alipay"
	}

	return baseURL + "/api/" + v + "/webhook/" + m
}
