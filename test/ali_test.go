package test

import (
	"encoding/json"
	"testing"
)

const aliResult = `{
	"alipay_trade_app_pay_response":{
		"code":"10000",
		"msg":"Success",
		"app_id":"2014072300007148",
		"out_trade_no":"081622560194853",
		"trade_no":"2016081621001004400236957647",
		"total_amount":"0.01",
		"seller_id":"2088702849871851",
		"charset":"utf-8",
		"timestamp":"2016-10-11 17:43:36"
	},
	"sign":"NGfStJf3i3ooWBuCDIQSumOpaGBcQz+aoAqyGh3W6EqA/gmyPYwLJ2REFijY9XPTApI9YglZyMw+ZMhd3kb0mh4RAXMrb6mekX4Zu8Nf6geOwIa9kLOnw0IMCjxi4abDIfXhxrXyj********",
	"sign_type":"RSA2"
}`

type alipayResult struct {
	Response map[string]string `json:"alipay_trade_app_pay_response"`
	Sign     string            `json:"sign"`
	SignTyp  string            `json:"sign_type"`
}

func TestMarshal(t *testing.T) {
	var r alipayResult

	err := json.Unmarshal([]byte(aliResult), &r)

	if err != nil {
		t.Error(err)
	}

	t.Logf("%+v\n", r)
}
