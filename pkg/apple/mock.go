// +build !production

package apple

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/guregu/null"
	"time"
)

type MockSubsBuilder struct {
	orgTxID string
	exp     time.Time
	edition price.Edition
	userID  string
}

// NewMockSubsBuilder creates a new instance.
// If `userID` is not empty, the subscription is linked to this id;
// otherwise the subscription is not linked to any ftc account.
func NewMockSubsBuilder(userID string) MockSubsBuilder {
	return MockSubsBuilder{
		orgTxID: faker.GenAppleSubID(),
		exp:     time.Now().AddDate(1, 0, 0),
		edition: price.StdYearEdition,
		userID:  userID,
	}
}

func (b MockSubsBuilder) WithEdition(e price.Edition) MockSubsBuilder {
	b.edition = e
	return b
}

func (b MockSubsBuilder) WithOriginalTxID(id string) MockSubsBuilder {
	b.orgTxID = id
	return b
}

func (b MockSubsBuilder) WithExpiration(t time.Time) MockSubsBuilder {
	b.exp = t
	return b
}

func (b MockSubsBuilder) Build() Subscription {
	return Subscription{
		BaseSchema: BaseSchema{
			Environment:           EnvSandbox,
			OriginalTransactionID: b.orgTxID,
		},
		LastTransactionID: faker.GenAppleSubID(),
		ProductID:         "",
		PurchaseDateUTC:   chrono.TimeNow(),
		ExpiresDateUTC:    chrono.TimeFrom(b.exp),
		Edition:           b.edition,
		AutoRenewal:       true,
		CreatedUTC:        chrono.TimeNow(),
		UpdatedUTC:        chrono.TimeNow(),
		FtcUserID:         null.NewString(b.userID, b.userID != ""),
		InUse:             true,
	}
}
