package reader

import (
	"testing"

	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/price"
)

func TestPaywall_Normalize(t *testing.T) {
	pw := Paywall{
		Products: []PaywallProduct{
			{
				Product: Product{
					Introductory: price.FtcPriceJSON{
						FtcPrice: price.MockFtcStdIntroPrice,
					},
				},
				Prices: []PaywallPrice{
					{
						FtcPrice: price.MockFtcStdYearPrice,
					},
					{
						FtcPrice: price.MockFtcStdMonthPrice,
					},
				},
			},
			{
				Prices: []PaywallPrice{
					{
						FtcPrice: price.MockFtcPrmPrice,
					},
				},
			},
		},
		Stripe: []StripePaywallItem{
			{
				Price: price.MockStripeStdIntroPrice,
			},
			{
				Price: price.MockStripeStdYearPrice,
			},
			{
				Price: price.MockStripeStdMonthPrice,
			},
			{
				Price: price.MockStripePrmPrice,
			},
		},
	}

	pw = pw.Normalize()

	t.Logf("%s", faker.MustMarshalIndent(pw))

	if pw.Products[0].Introductory.StripePriceID != price.MockStripeStdIntroPrice.ID {
		t.Errorf("got %s, expected %s", pw.Products[0].Introductory.StripePriceID, price.MockStripeStdIntroPrice.ID)
	}

	if pw.Products[0].Prices[0].StripePriceID != price.MockStripeStdYearPrice.ID {
		t.Errorf("got %s, expected %s", pw.Products[0].Prices[0].StripePriceID, price.MockStripeStdYearPrice.ID)
	}

	if pw.Products[0].Prices[1].StripePriceID != price.MockStripeStdMonthPrice.ID {
		t.Errorf("got %s, expected %s", pw.Products[0].Prices[1].StripePriceID, price.MockStripeStdMonthPrice.ID)
	}

	if pw.Products[1].Prices[0].StripePriceID != price.MockStripePrmPrice.ID {
		t.Errorf("got %s, expected %s", pw.Products[0].Prices[0].StripePriceID, price.MockStripePrmPrice.ID)
	}
}
