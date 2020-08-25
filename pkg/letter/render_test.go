package letter

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/product"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/brianvoe/gofakeit/v5"
	"github.com/guregu/null"
	"testing"
	"time"
)

func TestRenderNewSubs(t *testing.T) {
	faker.SeedGoFake()

	s, err := RenderNewSubs(CtxSubs{
		UserName: gofakeit.Username(),
		Order: subs.Order{
			ID: subs.MustGenerateOrderID(),
			Edition: product.Edition{
				Tier:  enum.TierStandard,
				Cycle: enum.CycleYear,
			},
			Charge: product.Charge{
				Amount:   128,
				Currency: "",
			},
			CreatedAt: chrono.TimeNow(),
			StartDate: chrono.DateNow(),
			EndDate:   chrono.DateFrom(time.Now().AddDate(1, 0, 0)),
		},
	})

	if err != nil {
		t.Error(err)
	}

	t.Logf("%s", s)
}

func TestRenderRenewalSubs(t *testing.T) {
	faker.SeedGoFake()

	s, err := RenderRenewalSubs(CtxSubs{
		UserName: gofakeit.Username(),
		Order: subs.Order{
			ID: subs.MustGenerateOrderID(),
			Edition: product.Edition{
				Tier:  enum.TierStandard,
				Cycle: enum.CycleYear,
			},
			Charge: product.Charge{
				Amount:   128,
				Currency: "",
			},
			CreatedAt: chrono.TimeNow(),
			StartDate: chrono.DateNow(),
			EndDate:   chrono.DateFrom(time.Now().AddDate(1, 0, 0)),
		},
	})

	if err != nil {
		t.Error(err)
	}

	t.Logf("%s", s)
}

func TestRenderUpgrade(t *testing.T) {
	faker.SeedGoFake()

	upID := subs.MustGenerateOrderID()

	s, err := RenderUpgrade(CtxUpgrade{
		UserName: gofakeit.Username(),
		Order: subs.Order{
			ID: subs.MustGenerateOrderID(),
			Edition: product.Edition{
				Tier:  enum.TierStandard,
				Cycle: enum.CycleYear,
			},
			Charge: product.Charge{
				Amount:   1000,
				Currency: "",
			},
			TotalBalance: null.FloatFrom(998),
			CreatedAt:    chrono.TimeNow(),
			StartDate:    chrono.DateNow(),
			EndDate:      chrono.DateFrom(time.Now().AddDate(1, 0, 0)),
		},
		Prorated: []subs.ProratedOrder{
			{
				OrderID:        subs.MustGenerateOrderID(),
				Balance:        50,
				CreatedUTC:     chrono.TimeNow(),
				ConsumedUTC:    chrono.TimeNow(),
				UpgradeOrderID: upID,
			},
			{
				OrderID:        subs.MustGenerateOrderID(),
				Balance:        21,
				CreatedUTC:     chrono.TimeNow(),
				ConsumedUTC:    chrono.TimeNow(),
				UpgradeOrderID: upID,
			},
		},
	})

	if err != nil {
		t.Error(err)
	}

	t.Logf("%s", s)
}
