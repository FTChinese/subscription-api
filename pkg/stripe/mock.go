package stripe

import (
	"encoding/json"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"github.com/stripe/stripe-go/v72"
)

func MockNewSubs() Subs {
	var ss stripe.Subscription
	if err := json.Unmarshal([]byte(faker.StripeSubs), &ss); err != nil {
		panic(err)
	}

	subs, err := NewSubs(&ss, reader.MemberID{
		CompoundID: "",
		FtcID:      null.StringFrom(uuid.New().String()),
		UnionID:    null.String{},
	}.MustNormalize())

	if err != nil {
		panic(err)
	}

	return subs
}
