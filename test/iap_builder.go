package test

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/apple"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/brianvoe/gofakeit/v5"
	"github.com/guregu/null"
	"time"
)

type IAPBuilder struct {
	env        apple.Environment
	origTxID   string
	lastTxID   string
	edition    price.Edition
	autoRenew  bool
	expiration time.Time
	ftcID      string
}

func NewIAPBuilder(origTxID string) IAPBuilder {
	if origTxID == "" {
		origTxID = faker.GenAppleSubID()
	}

	return IAPBuilder{
		env:        apple.EnvSandbox,
		origTxID:   origTxID,
		lastTxID:   faker.GenAppleSubID(),
		edition:    price.StdYearEdition,
		autoRenew:  true,
		expiration: time.Now().AddDate(1, 0, 0),
	}
}

func (b IAPBuilder) WithEnv(env apple.Environment) IAPBuilder {
	b.env = env
	return b
}

func (b IAPBuilder) WithSandbox() IAPBuilder {
	return b.WithEnv(apple.EnvProduction)
}

func (b IAPBuilder) WithProduction() IAPBuilder {
	return b.WithEnv(apple.EnvSandbox)
}

func (b IAPBuilder) WithLastTxID(id string) IAPBuilder {
	b.lastTxID = id
	return b
}

func (b IAPBuilder) WithEdition(e price.Edition) IAPBuilder {
	b.edition = e

	return b
}

func (b IAPBuilder) WithStdYear() IAPBuilder {
	return b.WithEdition(price.StdYearEdition)
}

func (b IAPBuilder) WithStdMonth() IAPBuilder {
	return b.WithEdition(price.StdMonthEdition)
}

func (b IAPBuilder) WithPrmYear() IAPBuilder {
	return b.WithEdition(price.PremiumEdition)
}

func (b IAPBuilder) WithAutoRenew(on bool) IAPBuilder {
	b.autoRenew = on
	return b
}

func (b IAPBuilder) WithAutoRenewOn() IAPBuilder {
	return b.WithAutoRenew(true)
}

func (b IAPBuilder) WithAutoRenewOff() IAPBuilder {
	return b.WithAutoRenew(false)
}

func (b IAPBuilder) WithExpiration(t time.Time) IAPBuilder {
	b.expiration = t
	return b
}

func (b IAPBuilder) WithFtcID(id string) IAPBuilder {
	b.ftcID = id
	return b
}

func (b IAPBuilder) ReceiptSchema() apple.ReceiptSchema {
	return apple.ReceiptSchema{
		BaseSchema: apple.BaseSchema{
			Environment:           b.env,
			OriginalTransactionID: b.origTxID,
		},
		LatestReceipt: gofakeit.Sentence(100),
		CreatedUTC:    chrono.TimeNow(),
	}
}

func (b IAPBuilder) Build() apple.Subscription {
	return apple.Subscription{
		BaseSchema: apple.BaseSchema{
			Environment:           b.env,
			OriginalTransactionID: b.origTxID,
		},
		LastTransactionID: b.lastTxID,
		ProductID:         "",
		PurchaseDateUTC:   chrono.TimeNow(),
		ExpiresDateUTC:    chrono.TimeFrom(b.expiration),
		Edition:           b.edition,
		AutoRenewal:       b.autoRenew,
		CreatedUTC:        chrono.TimeNow(),
		UpdatedUTC:        chrono.Time{},
		FtcUserID:         null.NewString(b.ftcID, b.ftcID != ""),
		InUse:             false,
	}
}
