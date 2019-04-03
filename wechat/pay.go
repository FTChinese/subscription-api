package wechat

import (
	"fmt"
	"github.com/FTChinese/go-rest"
	"time"
)

// Pay is a base struct for the response of wechat payment.
type Pay struct {
	FtcOrderID string  `json:"ftcOrderId"`
	ListPrice  float64 `json:"listPrice"`
	NetPrice   float64 `json:"netPrice"`
	AppID      string  `json:"appId"`
}

func GenerateNonce() string {
	nonce, _ := gorest.RandomHex(10)

	return nonce
}

func GenerateTimestamp() string {
	return fmt.Sprintf("%d", time.Now().Unix())
}
