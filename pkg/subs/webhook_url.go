package subs

import "github.com/FTChinese/go-rest/enum"

func WebhookURL(sandbox bool, method enum.PayMethod) string {
	var baseURL string
	if sandbox {
		baseURL = "https://www.ftacademy.cn/api/sandbox"
	} else {
		baseURL = "https://www.ftacademy.cn/api/v3"
	}

	switch method {
	case enum.PayMethodWx:
		return baseURL + "/webhook/wxpay"
	case enum.PayMethodAli:
		return baseURL + "/webhook/alipay"
	}

	return ""
}
