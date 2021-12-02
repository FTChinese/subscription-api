package stripeclient

import (
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/stripe"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestClient_ListPrices(t *testing.T) {
	faker.MustSetupViper()

	client := New(false, zaptest.NewLogger(t))
	stripePrices, err := client.ListPrices()
	if err != nil {
		t.Error(err)
		return
	}

	cache := stripe.NewPriceCache()
	cache.AddAll(stripePrices)

	prices := cache.List(false)

	for _, v := range prices {
		t.Logf("%s", faker.MustMarshalIndent(v))
	}
}

func TestClient_RetrievePrice(t *testing.T) {
	faker.MustSetupViper()

	client := New(false, zaptest.NewLogger(t))
	p, err := client.RetrievePrice("price_1Juuu2BzTK0hABgJTXiK4NTt")
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("%s", faker.MustMarshalIndent(p))
}
