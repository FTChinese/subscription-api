package striperepo

import (
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/repository/txrepo"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

// SubEnv wraps database connection
type StripeEnv struct {
	config.BuildConfig
	db *sqlx.DB
}

// NewSubEnv creates a new instance of SubEnv.
// `sandbox` is used to determine which table to write subscription data.
func NewStripeEnv(db *sqlx.DB, b config.BuildConfig) StripeEnv {
	return StripeEnv{
		BuildConfig: b,
		db:          db,
	}
}

func (env StripeEnv) beginAccountTx() (txrepo.AccountTx, error) {
	tx, err := env.db.Beginx()

	if err != nil {
		return txrepo.AccountTx{}, err
	}

	return txrepo.NewAccountTx(tx), nil
}

func (env StripeEnv) beginOrderTx() (txrepo.MemberTx, error) {
	tx, err := env.db.Beginx()

	if err != nil {
		return txrepo.MemberTx{}, err
	}

	return txrepo.NewMemberTx(tx), nil
}

var logger = logrus.
	WithField("project", "subscription-api").
	WithField("package", "stripe_repo")
