package repository

import (
	"errors"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/pw"
	"github.com/patrickmn/go-cache"
)

func paywallCacheKey(live bool) string {
	return "paywall_" + ids.GetBoolKey(live)
}

type CacheRepo struct {
	cache *cache.Cache
}

func NewCacheRepo(c *cache.Cache) CacheRepo {
	return CacheRepo{
		cache: c,
	}
}

func (repo CacheRepo) CachePaywall(p pw.Paywall, live bool) {
	repo.cache.Set(
		paywallCacheKey(live),
		p,
		cache.NoExpiration)
}

func (repo CacheRepo) LoadPaywall(live bool) (pw.Paywall, error) {
	x, found := repo.cache.Get(paywallCacheKey(live))
	if found {
		if paywall, ok := x.(pw.Paywall); ok {
			return paywall, nil
		}
	}

	return pw.Paywall{}, errors.New("not found")
}

func (repo CacheRepo) Clear() {
	repo.cache.Flush()
}
