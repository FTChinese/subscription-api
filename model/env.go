package model

import (
	"database/sql"

	"github.com/FTChinese/go-rest/enum"

	cache "github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
)

// Env wraps database connection
type Env struct {
	sandbox     bool
	memberTable string
	db          *sql.DB
	cache       *cache.Cache
}

// New creates a new instance of Env.
// `sandbox` is used to determine which table to write subscription data.
func New(db *sql.DB, c *cache.Cache, sandbox bool) Env {
	return Env{
		sandbox: sandbox,
		db:      db,
		cache:   c,
	}
}

// Get the database name used to store subscription and membership tables depending on the whether it is run in sandbox.
func (env Env) vipDBName() string {
	if env.sandbox {
		return "sandbox"
	}

	return "premium"
}

var logger = log.
	WithField("project", "subscription-api").
	WithField("package", "model")

const (
	keySchedule = "discountSchedule"
	keyPromo    = "promotionSchedule"
)

func normalizeMemberTier(vipType int64) enum.Tier {
	switch vipType {

	case 10:
		return enum.TierStandard

	case 100:
		return enum.TierPremium

	default:
		return enum.InvalidTier
	}
}
