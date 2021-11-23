package pw

import (
	"errors"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/guregu/null"
)

// PaywallDoc contains the banner for both daily user
// promotion. Each time the daily or promo version is edited,
// a new version is created and save to db, which means
// the data is immutable. We actually do not allow user to
// edit data.
// Only the last row is retrieved upon using.
type PaywallDoc struct {
	ID          int64       `json:"id" db:"id"` // Ignore this field when inserting data since it is auto-incremented.
	DailyBanner BannerJSON  `json:"banner" db:"daily_banner"`
	PromoBanner BannerJSON  `json:"promo" db:"promo_banner"`
	LiveMode    bool        `json:"liveMode" db:"live_mode"`
	CreatedUTC  chrono.Time `json:"createdUtc" db:"created_utc"`
}

// NewPaywallBase creates a new banner for paywall.
// The data is immutable. Every time a new row is created.
func NewPaywallBase(live bool) PaywallDoc {
	return PaywallDoc{
		ID:          0,
		DailyBanner: BannerJSON{},
		PromoBanner: BannerJSON{},
		LiveMode:    live,
		CreatedUTC:  chrono.TimeNow(),
	}
}

// WithBanner creates a new paywall by updating current paywall
// banner.
func (p PaywallDoc) WithBanner(b BannerJSON) PaywallDoc {
	p.DailyBanner = b
	p.CreatedUTC = chrono.TimeNow()

	return p
}

// WithPromo attaches a promotion banner to current paywall
// and creates a new version of PaywallDoc.
func (p PaywallDoc) WithPromo(b BannerJSON) PaywallDoc {
	p.PromoBanner = b
	p.CreatedUTC = chrono.TimeNow()

	return p
}

type Paywall struct {
	PaywallDoc
	Products []Product `json:"products"`
}

func NewPaywall(pwb PaywallDoc, p []Product) Paywall {
	return Paywall{
		PaywallDoc: pwb,
		Products:   p,
	}
}

func (w Paywall) findFtcPrice(id string) (price.FtcPrice, error) {
	for _, prod := range w.Products {
		for _, v := range prod.Prices {
			if v.ID == id {
				return v, nil
			}
		}
	}

	return price.FtcPrice{}, errors.New("the requested price is not found")
}

func (w Paywall) FindCheckoutItem(priceID string, offerID null.String) (price.CheckoutItem, error) {
	ftcPrice, err := w.findFtcPrice(priceID)
	if err != nil {
		return price.CheckoutItem{}, err
	}

	if offerID.IsZero() {
		return price.CheckoutItem{
			Price: ftcPrice.Price,
			Offer: price.Discount{},
		}, nil
	}

	offer, _ := ftcPrice.Offers.FindValid(offerID.String)

	return price.CheckoutItem{
		Price: ftcPrice.Price,
		Offer: offer,
	}, nil
}
