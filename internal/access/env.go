/*
Package access controls access right of a user or app to all API endpoints

When you are accessing the API directory from browser, add you access token to query parameter `access_key`;

When used by an app, set the token as the value of Bearer Token:

```
Authorization: Bearer <token>
```
*/
package access

import (
	"time"

	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/patrickmn/go-cache"
)

type Env struct {
	cache   *cache.Cache
	gormDBs db.MultiGormDBs
}

func NewEnv(dbs db.MultiGormDBs) Env {
	return Env{
		// Default expiration 24 hours, and purges the expired items every hour.
		cache:   cache.New(24*time.Hour, 1*time.Hour),
		gormDBs: dbs,
	}
}

// Load tries to load an access token from cache first, then
// retrieve from db if not found in cache.
func (env Env) Load(token string) (OAuth, error) {
	if acc, ok := env.loadCachedToken(token); ok {
		return acc, nil
	}

	acc, err := env.retrieveFromDB(token)
	if err != nil {
		return acc, err
	}

	env.cacheToken(token, acc)

	return acc, nil
}

func (env Env) loadCachedToken(token string) (OAuth, bool) {
	x, found := env.cache.Get(token)
	if !found {
		return OAuth{}, false
	}

	if access, ok := x.(OAuth); ok {
		return access, true
	}

	return OAuth{}, false
}

func (env Env) retrieveFromDB(token string) (OAuth, error) {
	var access OAuth

	err := env.gormDBs.Read.First(&access, "access_token = UNHEX(?)", token).Error

	if err != nil {
		return access, err
	}

	return access, nil
}

func (env Env) cacheToken(token string, access OAuth) {
	env.cache.Set(token, access, cache.DefaultExpiration)
}
