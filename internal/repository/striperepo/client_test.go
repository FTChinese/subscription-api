package striperepo

import (
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/brianvoe/gofakeit/v5"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestClient_CreateCustomer(t *testing.T) {
	client := NewClient(false, zaptest.NewLogger(t))

	cus, err := client.CreateCustomer(gofakeit.Email())

	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Customer %v", cus)
}

func TestClient_RetrieveCustomer(t *testing.T) {
	client := NewClient(false, zaptest.NewLogger(t))

	cus, err := client.RetrieveCustomer("cus_Ht6nKQQUq4ag2I")
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Customer %v", cus)
}

func TestClient_ListPrices(t *testing.T) {
	config.MustSetupViper()

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

	price.StripePriceCache.AddAll(stripePrices)

	prices := price.StripePriceCache.List(true)

	for _, sp := range prices {
		t.Logf("%s", faker.MustMarshalIndent(sp))
	}
}
