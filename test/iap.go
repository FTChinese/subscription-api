package test

import (
	"encoding/json"
	"github.com/Pallinder/go-randomdata"
	"gitlab.com/ftchinese/subscription-api/models/apple"
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

	return resp.LatestTransactions[randomdata.Number(l)]
}

func GetPendingRenewal() apple.PendingRenewal {
	resp := GetVerificationResponse()

	l := len(resp.PendingRenewalInfo)

	return resp.PendingRenewalInfo[randomdata.Number(l)]
}
