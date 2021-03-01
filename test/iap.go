// +build !production

package test

import (
	"encoding/json"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/rand"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/apple"
	"github.com/brianvoe/gofakeit/v5"
	"github.com/guregu/null"
	"time"
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

func (p *Persona) IAPSubs() apple.Subscription {
	s := apple.Subscription{
		BaseSchema: apple.BaseSchema{
			Environment:           apple.EnvSandbox,
			OriginalTransactionID: p.AppleSubID,
		},
		LastTransactionID: faker.GenAppleSubID(),
		ProductID:         "",
		PurchaseDateUTC:   chrono.TimeNow(),
		ExpiresDateUTC:    chrono.TimeFrom(time.Now().AddDate(1, 0, 0)),
		Edition:           p.price.Edition,
		AutoRenewal:       true,
	}

	if p.expired {
		s.ExpiresDateUTC = chrono.TimeFrom(time.Now().AddDate(-1, 0, 0))
	}

	return s
}

func (p *Persona) IAPSubsLinked() apple.Subscription {
	s := p.IAPSubs()
	s.FtcUserID = null.StringFrom(p.FtcID)

	return s
}

func (p *Persona) IAPLinkInput() apple.LinkInput {
	return apple.LinkInput{
		FtcID:        p.FtcID,
		OriginalTxID: p.AppleSubID,
	}
}

func (p *Persona) IAPReceiptSchema() apple.ReceiptSchema {
	return apple.ReceiptSchema{
		BaseSchema: apple.BaseSchema{
			Environment:           apple.EnvProduction,
			OriginalTransactionID: p.AppleSubID,
		},
		LatestReceipt: gofakeit.Sentence(100),
		CreatedUTC:    chrono.TimeNow(),
	}
}
