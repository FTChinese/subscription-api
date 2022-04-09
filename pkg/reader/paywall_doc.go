package reader

import "github.com/FTChinese/go-rest/chrono"

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

// NewPaywallDoc creates a new banner for paywall.
// The data is immutable. Every time a new row is created.
func NewPaywallDoc(live bool) PaywallDoc {
	return PaywallDoc{
		ID:          0,
		DailyBanner: BannerJSON{},
		PromoBanner: BannerJSON{},
		LiveMode:    live,
		CreatedUTC:  chrono.TimeNow(),
	}
}

func (p PaywallDoc) IsEmpty() bool {
	return p.ID > 0
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

func (p PaywallDoc) DropPromo() PaywallDoc {
	p.PromoBanner = BannerJSON{}
	p.CreatedUTC = chrono.TimeNow()

	return p
}
