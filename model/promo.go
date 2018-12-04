package model

import (
	"encoding/json"
	"time"

	cache "github.com/patrickmn/go-cache"
	"gitlab.com/ftchinese/subscription-api/util"
)

// Banner is the banner used on the barrier page
type Banner struct {
	CoverURL   string   `json:"coverUrl"`
	Heading    string   `json:"heading"`
	SubHeading string   `json:"subHeading"`
	Content    []string `json:"content"`
}

// Promotion contains a promotion's scheduled begining and ending time,
// pricing plans and barrier banner.
type Promotion struct {
	StartUTC  util.ISODateTime `json:"startAt"`
	EndUTC    util.ISODateTime `json:"endAt"`
	Plans     map[string]Plan  `json:"plans"`
	Banner    *Banner          `json:"banner"`
	CreatedAt string           `json:"createdAt"`
	createdBy string
}

// Test if now falls within the range of
// a promotion's start and end time.
func (p Promotion) isInEffect() bool {
	now := time.Now()
	start, err := p.StartUTC.ToTime()
	if err != nil {
		return false
	}
	end, err := p.EndUTC.ToTime()
	if err != nil {
		return false
	}

	// Start |------ now -------| End
	if now.Before(start) || now.After(end) {
		return false
	}
	return true
}

// RetrievePromo tries to retrieve a promotion schedule.
func (env Env) RetrievePromo() (Promotion, error) {
	// Retrieve a lastest promotion schedule
	// which is enabled,
	// ending time is equal to or greater than current time,
	// plans and banner columns are not null.
	// Since this is a promotion, plans must not be null;
	// otherwise the promotion is meaningless.
	// And banner must also not be null since as a promotion you should at least
	// tell user something about your promotion.
	query := `SELECT  
		start_utc AS startUtc,
		end_utc AS endUtc,
		plans AS plans,
		banner AS banner,
		created_utc AS createdUtc,
		created_by AS createdBy
	FROM premium.promotion_schedule
	WHERE end_utc >= UTC_TIMESTAMP()
		AND is_enabled = 1
		AND plans IS NOT NULL
		AND banner IS NOT NULL
	ORDER BY created_utc DESC
	LIMIT 1`

	var p Promotion
	var plans string
	var banner string

	err := env.DB.QueryRow(query).Scan(
		&p.StartUTC,
		&p.EndUTC,
		&plans,
		&banner,
		&p.CreatedAt,
		&p.createdBy,
	)

	// ErrNoRows
	if err != nil {
		logger.WithField("location", "RetrievePromo").Error(err)
		return p, err
	}

	if plans != "" {
		if err := json.Unmarshal([]byte(plans), &p.Plans); err != nil {
			return p, err
		}
	}

	if banner != "" {
		if err := json.Unmarshal([]byte(banner), &p.Banner); err != nil {
			return p, err
		}
	}

	env.cachePromo(p)

	return p, nil
}

func (env Env) cachePromo(p Promotion) {
	env.Cache.Set(keyPromo, p, cache.NoExpiration)
}

// PromoFromCache gets promo from cache.
func (env Env) PromoFromCache() (Promotion, bool) {
	if x, found := env.Cache.Get(keyPromo); found {
		promo, ok := x.(Promotion)

		if ok {
			logger.WithField("location", "PromoFromCache").Infof("Cached promo found %+v", promo)
			return promo, true
		}

		return Promotion{}, false
	}

	return Promotion{}, false
}
