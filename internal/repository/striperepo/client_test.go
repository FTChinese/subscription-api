package striperepo

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/pkg/product"
	"github.com/FTChinese/subscription-api/pkg/stripe"
	"github.com/brianvoe/gofakeit/v5"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestClient_GetPlan(t *testing.T) {

	client := NewClient(false, zaptest.NewLogger(t))

	p, err := client.GetPlan(product.Edition{
		Tier:  enum.TierStandard,
		Cycle: enum.CycleYear,
	})

	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("%v", p)
}

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
	stripePrices := client.ListPrices()

	t.Log(len(stripePrices))

	t.Log(stripePrices)

	for _, p := range stripePrices {
		t.Log(p)
	}

	stripe.PriceStore.AddAll(stripePrices)

	prices := stripe.PriceStore.List(true)

	for _, sp := range prices {
		t.Logf("%s", faker.MustMarshalIndent(sp))
	}
}
