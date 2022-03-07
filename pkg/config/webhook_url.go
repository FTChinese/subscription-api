package config

import "github.com/FTChinese/go-rest/enum"

const hostFTA = "https://www.ftacademy.cn"

const (
	// EmailVerificationURL is the base url to construct url in email verification letter.
	// Previously we used https://users.chineseft.com/verify/email created by the next-user app, which is no longer maintained.
	EmailVerificationURL = hostFTA + "/reader/verification"
	// PasswordResetURL is the base url to construct url to reset password.
	// Previously we used https://users.ftchinese.com/password-reset created by the next-user app.
	PasswordResetURL = hostFTA + "/reader/password-reset"
)

// AliWxWebhookURL builds the url for one-time purchase.
// If isProd is true, use online production server;
// otherwise always use sandbox server.
func AliWxWebhookURL(isProd bool, method enum.PayMethod) string {
	var baseURL string
	if isProd {
		baseURL = hostFTA + "/api/" + Version
	} else {
		baseURL = hostFTA + "/api/sandbox"
	}

	switch method {
	case enum.PayMethodWx:
		return baseURL + "/webhook/wxpay"
	case enum.PayMethodAli:
		return baseURL + "/webhook/alipay"
	}

	return ""
}
