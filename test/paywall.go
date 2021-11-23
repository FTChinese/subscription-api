package test

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/pw"
	"github.com/brianvoe/gofakeit/v5"
	"github.com/guregu/null"
	"time"
)

type ProductBuilder struct {
	productID string
	tier      enum.Tier
	live      bool
}

func NewProductBuilder(t enum.Tier) ProductBuilder {
	return ProductBuilder{
		productID: ids.ProductID(),
		tier:      t,
		live:      true,
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

func (b ProductBuilder) Build() pw.Product {
	faker.SeedGoFake()

	return pw.NewProduct(
		pw.ProductParams{
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
		live:   true,
		active: false,
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

func (b PriceBuilder) Build() price.Price {
	var amount float64
	if b.edition == price.StdMonthEdition {
		amount = 35
	} else if b.edition == price.StdYearEdition {
		amount = 298
	} else if b.edition == price.PremiumEdition {
		amount = 1998
	}

	return price.Price{
		ID:          b.priceID,
		Edition:     b.edition,
		Active:      b.active,
		Currency:    "cny",
		Description: null.String{},
		LiveMode:    b.live,
		Nickname:    null.String{},
		ProductID:   b.productID,
		UnitAmount:  amount,
		CreatedUTC:  chrono.TimeNow(),
		CreatedBy:   gofakeit.Username(),
	}
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
}

func (b DiscountBuilder) Build(k price.OfferKind) price.Discount {
	return price.Discount{
		ID: ids.DiscountID(),
		DiscountParams: price.DiscountParams{
			CreatedBy:   gofakeit.Username(),
			Description: null.StringFrom(gofakeit.Sentence(10)),
			Kind:        k,
			Percent:     null.Int{},
			DateTimePeriod: dt.DateTimePeriod{
				StartUTC: chrono.TimeNow(),
				EndUTC:   chrono.TimeFrom(time.Now().AddDate(0, 0, 7)),
			},
			PriceOff:  null.FloatFrom(b.off),
			PriceID:   b.priceID,
			Recurring: false,
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
