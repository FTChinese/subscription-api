package wxpay

import (
	"fmt"
	"github.com/FTChinese/go-rest/rand"
	"time"
)

func NonceStr() string {
	return rand.String(10)
}

func GenerateTimestamp() string {
	return fmt.Sprintf("%d", time.Now().Unix())
}
