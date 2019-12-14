package striperepo

import (
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"gitlab.com/ftchinese/subscription-api/models/util"
	"gitlab.com/ftchinese/subscription-api/repository/query"
)

// SubEnv wraps database connection
type StripeEnv struct {
	util.BuildConfig
	db *sqlx.DB
}

// NewSubEnv creates a new instance of SubEnv.
// `sandbox` is used to determine which table to write subscription data.
func NewStripeEnv(db *sqlx.DB, b util.BuildConfig) StripeEnv {
	return StripeEnv{
		BuildConfig: b,
		db:          db,
	}
}

func (env StripeEnv) beginAccountTx() (query.AccountTx, error) {
	tx, err := env.db.Beginx()

	if err != nil {
		return query.AccountTx{}, err
	}

	return query.NewAccountTx(tx), nil
}

func (env StripeEnv) beginOrderTx() (query.OrderTx, error) {
	tx, err := env.db.Beginx()

	if err != nil {
		return query.OrderTx{}, err
	}

	return query.NewOrderTx(tx, env.UseSandboxDB()), nil
}

var logger = logrus.
	WithField("project", "subscription-api").
	WithField("package", "stripe_repo")
