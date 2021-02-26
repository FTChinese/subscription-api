package pw

import (
	"github.com/FTChinese/subscription-api/lib/dt"
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
}

func NewPaywall(b BannerSchema, p []Product) Paywall {
	return Paywall{
		Banner:   b.Banner,
		Promo:    b.Promo,
		Products: p,
	}
}
