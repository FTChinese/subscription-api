//go:build !production
// +build !production

package test

import (
	"crypto"
	"encoding/base64"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/subscription-api/internal/pkg/subs"
	"github.com/smartwalle/alipay"
	"github.com/smartwalle/alipay/encoding"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

type AliPayWebhook struct {
	NotifyTime    string `schema:"notify_time"`
	NotifyType    string `schema:"notify_type"`
	NotifyID      string `schema:"notify_id"`
	AppID         string `schema:"app_id"`
	Charset       string `schema:"charset"`
	Version       string `schema:"version"`
	SignType      string `schema:"sign_type"`
	TradeNo       string `schema:"trade_no"`
	OutTradeNo    string `schema:"out_trade_no"`
	TradeStatus   string `schema:"trade_status"`
	TotalAmount   string `schema:"total_amount"`
	ReceiptAmount string `schema:"receipt_amount"`
}

func NewAlipayWebhook(order subs.Order) AliPayWebhook {
	return AliPayWebhook{
		NotifyTime:    chrono.TimeNow().Format(chrono.SQLDateTime),
		NotifyType:    "trade_status_sync",
		NotifyID:      "ac05699524730693a8b330c5ecf72da9786",
		AppID:         "2018072300007148",
		Charset:       "utf-8",
		Version:       "1.0",
		SignType:      "RSA2",
		TradeNo:       "2018112011001004330000121536",
		OutTradeNo:    order.ID,
		TradeStatus:   "TRADE_SUCCESS",
		TotalAmount:   order.AliPayable(),
		ReceiptAmount: order.AliPayable(),
	}
}

func (w AliPayWebhook) URLValues() url.Values {
	form := url.Values{}

	err := encoder.Encode(w, form)
	if err != nil {
		panic(err)
	}

	return form
}

func (w AliPayWebhook) SignedParams() (url.Values, error) {
	p := w.URLValues()

	var hash crypto.Hash

	if w.SignType == "RSA" {
		hash = crypto.SHA1
	} else {
		hash = crypto.SHA256
	}

	sign, err := generateAliSign(
		p,
		encoding.FormatPrivateKey(AliApp.PrivateKey),
		hash)

	if err != nil {
		return nil, err
	}

	p.Add("sign", sign)

	return p, nil
}

func (w AliPayWebhook) Encode() string {

	p, err := w.SignedParams()

	if err != nil {
		panic(err)
	}

	return p.Encode()
}

func generateAliSign(param url.Values, privateKey []byte, hash crypto.Hash) (string, error) {
	list := make([]string, 0)

	for k := range param {
		if k == "sign_type" || k == "sign" {
			continue
		}

		v := param.Get(k)
		if len(v) > 0 {
			list = append(list, k+"="+v)
		}
	}

	sort.Strings(list)

	src := strings.Join(list, "&")

	sig, err := encoding.SignPKCS1v15([]byte(src), privateKey, hash)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(sig), nil
}

func GetMockPayload(req *http.Request) *alipay.TradeNotification {
	p := &alipay.TradeNotification{}

	p.NotifyTime = req.FormValue("notify_time")
	p.NotifyType = req.FormValue("notify_type")
	p.NotifyId = req.FormValue("notify_id")
	p.AppId = req.FormValue("app_id")
	p.Charset = req.FormValue("charset")
	p.Version = req.FormValue("version")
	p.SignType = req.FormValue("sign_type")
	p.Sign = req.FormValue("sign")
	p.TradeNo = req.FormValue("trade_no")
	p.OutTradeNo = req.FormValue("out_trade_no")
	p.TradeStatus = req.FormValue("trade_status")
	p.TotalAmount = req.FormValue("total_amount")
	p.ReceiptAmount = req.FormValue("receipt_amount")

	return p
}
