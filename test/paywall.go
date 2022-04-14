//go:build !production
// +build !production

package test

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/brianvoe/gofakeit/v5"
	"github.com/guregu/null"
	"time"
)

func NewDailyBanner() reader.BannerJSON {
	faker.SeedGoFake()

	return reader.BannerJSON{
		ID:           ids.BannerID(),
		Heading:      gofakeit.Word(),
		SubHeading:   null.StringFrom(gofakeit.Sentence(5)),
		CoverURL:     null.StringFrom(gofakeit.URL()),
		Content:      null.StringFrom(gofakeit.Paragraph(2, 1, 10, "\n")),
		Terms:        null.String{},
		ChronoPeriod: dt.ChronoPeriod{},
	}
}

func NewPromoBanner() reader.BannerJSON {
	b := NewDailyBanner()

	b.ChronoPeriod = dt.ChronoPeriod{
		StartUTC: chrono.TimeNow(),
		EndUTC:   chrono.TimeFrom(time.Now().AddDate(0, 0, 7)),
	}

	return b
}

type ProductBuilder struct {
	productID string
	tier      enum.Tier
	live      bool
}

func NewProductBuilder(t enum.Tier) ProductBuilder {
	return ProductBuilder{
		productID: ids.ProductID(),
		tier:      t,
		live:      false,
	}
}

func NewStdProdBuilder() ProductBuilder {
	return NewProductBuilder(enum.TierStandard)
}

func NewPrmProdBuilder() ProductBuilder {
	return NewProductBuilder(enum.TierPremium)
}

func (b ProductBuilder) WithLive() ProductBuilder {
	b.live = true
	return b
}

func (b ProductBuilder) WithSandbox() ProductBuilder {
	b.live = false
	return b
}

func (b ProductBuilder) Build() reader.Product {
	faker.SeedGoFake()

	return reader.NewProduct(
		reader.ProductParams{
			CreatedBy: gofakeit.Username(),
			Description: null.StringFrom(gofakeit.Paragraph(
				4, 1, 10, "\n")),
			Heading:    b.tier.StringCN(),
			SmallPrint: null.String{},
			Tier:       b.tier,
		},
		b.live,
	)
}

func (b ProductBuilder) NewPriceBuilder(c enum.Cycle) PriceBuilder {
	return PriceBuilder{
		productID: b.productID,
		priceID:   ids.PriceID(),
		edition: price.Edition{
			Tier:  b.tier,
			Cycle: c,
		},
		live:   false,
		active: false,
		kind:   price.KindRecurring,
	}
}

func (b ProductBuilder) NewYearPriceBuilder() PriceBuilder {
	return b.NewPriceBuilder(enum.CycleYear)
}

func (b ProductBuilder) NewMonthPriceBuilder() PriceBuilder {
	return b.NewPriceBuilder(enum.CycleMonth)
}

type PriceBuilder struct {
	productID string
	priceID   string
	edition   price.Edition
	live      bool
	active    bool
	kind      price.Kind
}

func (b PriceBuilder) WithLive() PriceBuilder {
	b.live = true
	return b
}

func (b PriceBuilder) WithTest() PriceBuilder {
	b.live = false
	return b
}

func (b PriceBuilder) WithActive() PriceBuilder {
	b.active = true
	return b
}

func (b PriceBuilder) WithOneTime() PriceBuilder {
	b.kind = price.KindOneTime
	return b
}

func (b PriceBuilder) WithRecurring() PriceBuilder {
	b.kind = price.KindRecurring
	return b
}

func (b PriceBuilder) Build() price.FtcPrice {
	var amount float64
	if b.edition == price.StdMonthEdition {
		amount = 35
	} else if b.edition == price.StdYearEdition {
		amount = 298
	} else if b.edition == price.PremiumEdition {
		amount = 1998
	}

	p := price.New(price.FtcCreationParams{
		Kind:    b.kind,
		Edition: b.edition,
		FtcUpdateParams: price.FtcUpdateParams{
			Title:    null.StringFrom(gofakeit.Sentence(3)),
			Nickname: null.StringFrom(gofakeit.Sentence(2)),
			PeriodCount: price.ColumnYearMonthDay{
				YearMonthDay: dt.NewYearMonthDay(b.edition.Cycle),
			},
			StripePriceID: faker.StripePriceID(),
		},

		ProductID:  b.productID,
		UnitAmount: amount,
	}, b.live)

	if b.active {
		return p.Activate()
	}

	return p
}

func (b PriceBuilder) NewDiscountBuilder() DiscountBuilder {
	var off float64
	if b.edition == price.StdMonthEdition {
		off = 34
	} else if b.edition == price.StdYearEdition {
		off = 50
	} else if b.edition == price.PremiumEdition {
		off = 100
	}

	return DiscountBuilder{
		priceID: b.priceID,
		off:     off,
		live:    b.live,
	}
}

type DiscountBuilder struct {
	priceID string
	off     float64
	live    bool
	period  dt.YearMonthDay
}

func NewDiscountBuilder(priceID string) DiscountBuilder {
	return DiscountBuilder{
		priceID: priceID,
		off:     0,
		live:    false,
		period:  dt.YearMonthDay{},
	}
}

func (b DiscountBuilder) WithPriceOff(off float64) DiscountBuilder {
	b.off = off

	return b
}

func (b DiscountBuilder) WithMode(live bool) DiscountBuilder {
	b.live = live

	return b
}

func (b DiscountBuilder) WithPeriod(p dt.YearMonthDay) DiscountBuilder {
	b.period = p

	return b
}

func (b DiscountBuilder) Build(k price.OfferKind) price.Discount {
	return price.Discount{
		ID: ids.DiscountID(),
		DiscountParams: price.DiscountParams{
			Description: null.StringFrom(gofakeit.Sentence(10)),
			Kind:        k,
			OverridePeriod: price.ColumnYearMonthDay{
				YearMonthDay: b.period,
			},
			Percent:   null.Int{},
			PriceOff:  null.FloatFrom(b.off),
			PriceID:   b.priceID,
			Recurring: false,
			ChronoPeriod: dt.ChronoPeriod{
				StartUTC: chrono.TimeNow(),
				EndUTC:   chrono.TimeFrom(time.Now().AddDate(0, 0, 7)),
			},
			CreatedBy: gofakeit.Username(),
		},
		LiveMode:   b.live,
		Status:     price.DiscountStatusActive,
		CreatedUTC: chrono.TimeNow(),
	}
}

func (b DiscountBuilder) BuildIntro() price.Discount {
	return b.Build(price.OfferKindIntroductory)
}

func (b DiscountBuilder) BuildPromo() price.Discount {
	return b.Build(price.OfferKindPromotion)
}
func (b DiscountBuilder) BuildRetention() price.Discount {
	return b.Build(price.OfferKindRetention)
}
func (b DiscountBuilder) BuildWinBack() price.Discount {
	return b.Build(price.OfferKindWinBack)
}
