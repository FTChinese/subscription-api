package repository

import (
	"errors"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/reader"
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

func (repo CacheRepo) CachePaywall(p reader.Paywall, live bool) {
	repo.cache.Set(
		paywallCacheKey(live),
		p,
		cache.NoExpiration)
}

func (repo CacheRepo) LoadPaywall(live bool) (reader.Paywall, error) {
	x, found := repo.cache.Get(paywallCacheKey(live))
	if found {
		if paywall, ok := x.(reader.Paywall); ok {
			return paywall, nil
		}
	}

	return reader.Paywall{}, errors.New("not found")
}

func (repo CacheRepo) Clear() {
	repo.cache.Flush()
}
