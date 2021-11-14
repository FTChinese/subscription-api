package striperepo

import (
	"github.com/FTChinese/subscription-api/faker"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestClient_ListPrices(t *testing.T) {
	faker.MustSetupViper()

	client := NewClient(false, zaptest.NewLogger(t))
	stripePrices, err := client.ListPrices()
	if err != nil {
		t.Error(err)
		return
	}

	PriceCache.AddAll(stripePrices)

	prices := PriceCache.List(false)

	for _, v := range prices {
		t.Logf("%s", faker.MustMarshalIndent(v))
	}
}
