package pw

import (
	"errors"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/guregu/null"
)

type Banner struct {
	ID         int64       `json:"id" db:"banner_id"`
	Heading    string      `json:"heading" db:"heading"`
	SubHeading null.String `json:"subHeading" db:"sub_heading"`
	CoverURL   null.String `json:"coverUrl" db:"cover_url"`
	Content    null.String `json:"content" db:"content"`
}

type Promo struct {
	PromoID    null.String `json:"id" db:"promo_id"`
	Heading    null.String `json:"heading" db:"promo_heading"`
	SubHeading null.String `json:"subHeading" db:"promo_sub_heading"`
	CoverURL   null.String `json:"coverUrl" db:"promo_cover_url"`
	Content    null.String `json:"content" db:"promo_content"`
	Terms      null.String `json:"terms" db:"terms_conditions"`
	dt.DateTimePeriod
}

// BannerSchema represents data when retrieving banner by joining promo.
type BannerSchema struct {
	Banner
	Promo
}

type Paywall struct {
	Banner   Banner    `json:"banner"`
	Promo    Promo     `json:"promo"`
	Products []Product `json:"products"`
	LiveMode bool      `json:"liveMode"`
}

func NewPaywall(b BannerSchema, p []Product, live bool) Paywall {
	return Paywall{
		Banner:   b.Banner,
		Promo:    b.Promo,
		Products: p,
		LiveMode: live,
	}
}

func (w Paywall) CollectPrices() []price.FtcPrice {
	var list = make([]price.FtcPrice, 0)
	for _, product := range w.Products {
		for _, p := range product.Prices {
			list = append(list, p)
		}
	}
	return list
}

func (w Paywall) FindPrice(p price.Price) (price.FtcPrice, error) {
	for _, prod := range w.Products {
		for _, v := range prod.Prices {
			if v.ID == p.ID {
				return v, nil
			}
		}
	}

	return price.FtcPrice{}, errors.New("the requested price is not found")
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

func (w Paywall) FindCheckoutItem(priceID, offerID string) (price.CheckoutItem, error) {
	ftcPrice, err := w.findFtcPrice(priceID)
	if err != nil {
		return price.CheckoutItem{}, err
	}

	offer, _ := ftcPrice.Offers.FindValid(offerID)

	return price.CheckoutItem{
		Price: ftcPrice.Price,
		Offer: offer,
	}, nil
}
