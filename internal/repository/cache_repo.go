package repository

import (
	"errors"
	"github.com/FTChinese/subscription-api/internal/pkg/android"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/patrickmn/go-cache"
)

func paywallCacheKey(live bool) string {
	return "paywall_" + ids.GetBoolKey(live)
}

const cacheKeyAndroidLatest = "android_latest_release"

type CacheRepo struct {
	cache *cache.Cache
}

func NewCacheRepo(c *cache.Cache) CacheRepo {
	return CacheRepo{
		cache: c,
	}
}

// CachePaywall saves paywall data in cache.
// It never expires.
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

func (repo CacheRepo) AndroidLatest(release android.Release) {
	repo.cache.Set(
		cacheKeyAndroidLatest,
		release,
		cache.DefaultExpiration)
}

func (repo CacheRepo) LoadAndroidLatest() (android.Release, error) {
	x, found := repo.cache.Get(cacheKeyAndroidLatest)
	if found {
		if release, ok := x.(android.Release); ok {
			return release, nil
		}
	}

	return android.Release{}, errors.New("not found")
}

func (repo CacheRepo) Clear() {
	repo.cache.Flush()
}
