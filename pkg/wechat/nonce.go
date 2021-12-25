package wechat

import (
	"fmt"
	gorest "github.com/FTChinese/go-rest"
	"time"
)

func GenerateNonce() string {
	nonce, _ := gorest.RandomHex(10)

	return nonce
}

func GenerateTimestamp() string {
	return fmt.Sprintf("%d", time.Now().Unix())
}
