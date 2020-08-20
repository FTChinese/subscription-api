package test

import (
	"encoding/json"
	"github.com/FTChinese/go-rest/rand"
	"github.com/FTChinese/subscription-api/pkg/apple"
)

func GetReceiptToken() string {
	return iapReceipt
}

func GetVerificationResponse() *apple.VerificationResponseBody {
	var r apple.VerificationResponseBody
	if err := json.Unmarshal([]byte(iapVerificationResponse), &r); err != nil {
		panic(err)
	}

	return &r
}

func GetIAPTransaction() apple.Transaction {
	resp := GetVerificationResponse()

	l := len(resp.LatestTransactions)

	return resp.LatestTransactions[rand.IntRange(0, l)]
}

func GetPendingRenewal() apple.PendingRenewal {
	resp := GetVerificationResponse()

	l := len(resp.PendingRenewalInfo)

	return resp.PendingRenewalInfo[rand.IntRange(0, l)]
}
