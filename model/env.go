package model

import (
	"database/sql"
	"github.com/sirupsen/logrus"
	"gitlab.com/ftchinese/subscription-api/query"

	"github.com/FTChinese/go-rest/enum"

	"github.com/patrickmn/go-cache"
)

// Env wraps database connection
type Env struct {
	sandbox bool
	db      *sql.DB
	cache   *cache.Cache
	query   query.Builder
}

// New creates a new instance of Env.
// `sandbox` is used to determine which table to write subscription data.
func New(db *sql.DB, c *cache.Cache, sandbox bool) Env {
	return Env{
		sandbox: sandbox,
		db:      db,
		cache:   c,
		query:   query.NewBuilder(sandbox),
	}
}

// BeginMemberTx creates a MemberTx instance.
func (env Env) BeginMemberTx() (MemberTx, error) {
	tx, err := env.db.Begin()

	if err != nil {
		return MemberTx{}, err
	}

	return MemberTx{tx: tx, query: env.query}, nil
}

func (env Env) BeginOrderTx() (OrderTx, error) {
	tx, err := env.db.Begin()

	if err != nil {
		return OrderTx{}, err
	}

	return OrderTx{tx: tx, query: env.query}, nil
}

func (env Env) BeginStripeTx() (StripeTx, error) {
	tx, err := env.db.Begin()

	if err != nil {
		return StripeTx{}, err
	}

	return StripeTx{tx: tx, query: env.query}, nil
}

var logger = logrus.
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
