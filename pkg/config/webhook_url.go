package config

import "github.com/FTChinese/go-rest/enum"

// AliWxWebhookURL builds the url for one-time purchase.
// If isProd is true, use online production server;
// otherwise always use sandbox server.
func AliWxWebhookURL(isProd bool, method enum.PayMethod) string {
	var baseURL string
	if isProd {
		baseURL = aliWxWebhookBaseURL
	} else {
		baseURL = "https://www.ftacademy.cn/api/sandbox"
	}

	switch method {
	case enum.PayMethodWx:
		return baseURL + "/webhook/wxpay"
	case enum.PayMethodAli:
		return baseURL + "/webhook/alipay"
	}

	return ""
}
