package model

import (
	"encoding/json"

	cache "github.com/patrickmn/go-cache"
	"gitlab.com/ftchinese/subscription-api/paywall"
)

// RetrievePromo tries to retrieve a promotion schedule.
func (env Env) RetrievePromo() (paywall.Promotion, error) {
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
		created_utc AS createdUtc
	FROM premium.promotion_schedule
	WHERE end_utc >= UTC_TIMESTAMP()
		AND is_enabled = 1
		AND plans IS NOT NULL
		AND banner IS NOT NULL
	ORDER BY created_utc DESC
	LIMIT 1`

	var p paywall.Promotion
	var plans string
	var banner string

	err := env.db.QueryRow(query).Scan(
		&p.StartUTC,
		&p.EndUTC,
		&plans,
		&banner,
		&p.CreatedAt,
	)

	// ErrNoRows
	if err != nil {
		logger.WithField("trace", "RetrievePromo").Error(err)
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

func (env Env) cachePromo(p paywall.Promotion) {
	env.cache.Set(keyPromo, p, cache.NoExpiration)
}

// LoadCachedPromo gets promo from cache.
func (env Env) LoadCachedPromo() (paywall.Promotion, bool) {
	x, found := env.cache.Get(keyPromo)

	if !found {
		return paywall.Promotion{}, false
	}

	if promo, ok := x.(paywall.Promotion); ok {
		logger.WithField("trace", "LoadCachedPromo").Infof("Cached promo found %+v", promo)

		return promo, true
	}

	return paywall.Promotion{}, false
}

// GetCurrentPricing get current effective pricing plans.
func (env Env) GetCurrentPricing() paywall.Pricing {
	if env.sandbox {
		return paywall.GetSandboxPricing()
	}

	promo, found := env.LoadCachedPromo()
	if !found {
		logger.WithField("trace", "GetCurrentPricing").Info("Promo not found. Use default plans")

		return paywall.GetDefaultPricing()
	}
	if !promo.IsInEffect() {
		logger.WithField("trace", "GetCurrentPricing").Info("Promo is not in effetive time range. Use default plans")

		return paywall.GetDefaultPricing()
	}

	logger.WithField("trace", "GetCurrentPricing").Info("Use promotion pricing plans")

	return promo.Plans
}
