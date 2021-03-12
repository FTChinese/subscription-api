package subs

import "github.com/FTChinese/go-rest/enum"

func WebhookURL(sandbox bool, method enum.PayMethod) string {
	var baseURL string
	if sandbox {
		baseURL = "http://www.ftacademy.cn/api/sandbox"
	} else {
		baseURL = "http://www.ftacademy.cn/api/v2"
	}

	switch method {
	case enum.PayMethodWx:
		return baseURL + "/webhook/wxpay"
	case enum.PayMethodAli:
		return baseURL + "/webhook/alipay"
	}

	return ""
}
