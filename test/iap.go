//go:build !production
// +build !production

package test

import (
	"encoding/json"
	"github.com/FTChinese/go-rest/rand"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/apple"
)

func MustVerificationResponse() *apple.VerificationResp {
	var r apple.VerificationResp
	if err := json.Unmarshal([]byte(faker.IAPVerificationResponse), &r); err != nil {
		panic(err)
	}

	r.Parse()

	return &r
}

func MustIAPTransaction() apple.Transaction {
	resp := MustVerificationResponse()

	l := len(resp.LatestReceiptInfo)

	return resp.LatestReceiptInfo[rand.IntRange(0, l)]
}

func MustPendingRenewal() apple.PendingRenewal {
	resp := MustVerificationResponse()

	l := len(resp.PendingRenewalInfo)

	return resp.PendingRenewalInfo[rand.IntRange(0, l)]
}
