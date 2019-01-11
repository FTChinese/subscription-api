package paywall

import (
	cache "github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
)

var logger = log.WithField("package", "paywall")

// Env wraps database connection
type Env struct {
	Cache *cache.Cache
}
