package test

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/rand"
	"github.com/smartwalle/alipay"
	"time"
)

func AliNoti() alipay.TradeNotification {
	return alipay.TradeNotification{
		NotifyTime: time.Now().In(time.UTC).Format(chrono.SQLDateTime),
		NotifyType: "trade_status_sync",
		NotifyId:   rand.String(36),
		AppId:      AliApp.ID,
		Charset:    "utf-8",
		Version:    "1.0",
		SignType:   "RSA2",
		Sign:       rand.String(256),
		TradeNo:    rand.String(64),
		OutTradeNo: rand.String(18),
		GmtCreate:  time.Now().In(time.UTC).Format(chrono.SQLDateTime),
		GmtPayment: time.Now().In(time.UTC).Format(chrono.SQLDateTime),
	}
}
