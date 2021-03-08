package apple

import (
	"github.com/FTChinese/subscription-api/pkg"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSubscription_NewMembership(t *testing.T) {
	resp := mustParsedReceiptResponse()
	sub, err := NewSubscription(resp.UnifiedReceipt)
	if err != nil {
		t.Error(err)
	}

	got := NewMembership(pkg.MemberID{
		CompoundID: "",
		FtcID:      null.StringFrom(uuid.New().String()),
		UnionID:    null.String{},
	}.MustNormalize(), sub)

	got = got.Sync()

	t.Logf("%+v", got)

	assert.NotZero(t, got.LegacyTier)
	assert.NotZero(t, got.LegacyExpire)
}

func TestSubscription_BuildOn(t *testing.T) {
	resp := mustParsedReceiptResponse()
	sub, err := NewSubscription(resp.UnifiedReceipt)
	if err != nil {
		t.Error(err)
	}

	got := NewMembership(pkg.MemberID{
		CompoundID: "",
		FtcID:      null.StringFrom(uuid.New().String()),
		UnionID:    null.String{},
	}.MustNormalize(), sub)

	got = got.Sync()

	t.Logf("%+v", got)

	assert.NotZero(t, got.LegacyTier)
	assert.NotZero(t, got.LegacyExpire)
}
