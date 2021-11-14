package striperepo

import (
	"github.com/FTChinese/subscription-api/faker"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestClient_ListPrices(t *testing.T) {
	faker.MustSetupViper()

	client := NewClient(true, zaptest.NewLogger(t))
	stripePrices, err := client.ListPrices()
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(len(stripePrices))

	t.Log(stripePrices)

	for _, p := range stripePrices {
		t.Log(p)
	}

	PriceCache.AddAll(stripePrices)

	prices := PriceCache.List(true)

	for _, sp := range prices {
		t.Logf("%s", faker.MustMarshalIndent(sp))
	}
}
