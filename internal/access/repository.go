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
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/patrickmn/go-cache"
	"time"
)

type Repo struct {
	dbs   db.ReadWriteSplit
	cache *cache.Cache
}

func NewRepo(dbs db.ReadWriteSplit) Repo {
	return Repo{
		dbs: dbs,
		// Default expiration 24 hours, and purges the expired items every hour.
		cache: cache.New(24*time.Hour, 1*time.Hour),
	}
}

// Load tries to load an access token from cache first, then
// retrieve from db if not found in cache.
func (env Repo) Load(token string) (OAuthAccess, error) {
	if acc, ok := env.LoadCachedToken(token); ok {
		return acc, nil
	}

	acc, err := env.RetrieveFromDB(token)
	if err != nil {
		return acc, err
	}

	env.CacheToken(token, acc)

	return acc, nil
}

func (env Repo) RetrieveFromDB(token string) (OAuthAccess, error) {
	var access OAuthAccess

	if err := env.dbs.Read.Get(&access, stmtOAuth, token); err != nil {
		return access, err
	}

	return access, nil
}

func (env Repo) CacheToken(token string, access OAuthAccess) {
	env.cache.Set(token, access, cache.DefaultExpiration)
}

func (env Repo) LoadCachedToken(token string) (OAuthAccess, bool) {
	x, found := env.cache.Get(token)
	if !found {
		return OAuthAccess{}, false
	}

	if access, ok := x.(OAuthAccess); ok {
		return access, true
	}

	return OAuthAccess{}, false
}
