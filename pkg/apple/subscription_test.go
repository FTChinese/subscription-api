package apple

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/product"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSubscription_NewMembership(t *testing.T) {
	resp := mustParsedReceiptResponse()
	sub, err := resp.Subscription()
	if err != nil {
		t.Error(err)
	}

	got := sub.NewMembership(reader.MemberID{
		CompoundID: "",
		FtcID:      null.StringFrom(uuid.New().String()),
		UnionID:    null.String{},
	}.MustNormalize())

	got = got.Sync()

	t.Logf("%+v", got)

	assert.NotZero(t, got.LegacyTier)
	assert.NotZero(t, got.LegacyExpire)
}

func TestSubscription_BuildOn(t *testing.T) {
	resp := mustParsedReceiptResponse()
	sub, err := resp.Subscription()
	if err != nil {
		t.Error(err)
	}

	got := sub.BuildOn(reader.Membership{
		MemberID: reader.MemberID{
			CompoundID: "",
			FtcID:      null.StringFrom(uuid.New().String()),
			UnionID:    null.String{},
		}.MustNormalize(),
		Edition: product.Edition{
			Tier:  enum.TierStandard,
			Cycle: enum.CycleYear,
		},
		LegacyTier:    null.Int{},
		LegacyExpire:  null.Int{},
		ExpireDate:    chrono.DateFrom(time.Now().AddDate(-1, 0, 0)),
		PaymentMethod: enum.PayMethodAli,
		FtcPlanID:     null.String{},
		StripeSubsID:  null.String{},
		StripePlanID:  null.String{},
		AutoRenewal:   false,
		Status:        0,
		AppleSubsID:   null.String{},
		B2BLicenceID:  null.String{},
	})

	got = got.Sync()

	t.Logf("%+v", got)

	assert.NotZero(t, got.LegacyTier)
	assert.NotZero(t, got.LegacyExpire)
}
