package iaprepo

import (
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/repository/txrepo"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

var logger = logrus.
	WithField("project", "subscription-api").
	WithField("package", "iap_repo")

type Env struct {
	cfg config.BuildConfig
	db  *sqlx.DB
}

func NewEnv(db *sqlx.DB, c config.BuildConfig) Env {
	return Env{
		cfg: c,
		db:  db,
	}
}

// BeginTx starts a transaction.
// NOTE: here the sandbox is different from the environment
// field send by apple. It only determines whether the
// sandbox db should be used and is determined by
// the CLI argument `-sandbox`.
// All messages from apple is save in production DBs.
func (env Env) BeginTx() (txrepo.MemberTx, error) {
	tx, err := env.db.Beginx()

	if err != nil {
		return txrepo.MemberTx{}, err
	}

	return txrepo.NewMemberTx(tx), nil
}
