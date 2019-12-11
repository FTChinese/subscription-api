package iaprepo

import (
	"errors"
	"github.com/jmoiron/sqlx"
	"github.com/parnurzeal/gorequest"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gitlab.com/ftchinese/subscription-api/models/util"
)

var logger = logrus.
	WithField("project", "subscription-api").
	WithField("package", "iaprepo")

var request = gorequest.New()

func getReceiptPassword() (string, error) {
	pw := viper.GetString("apple.receipt_password")
	if pw == "" {
		return "", errors.New("empty receipt verification password")
	}

	return pw, nil
}

type IAPEnv struct {
	c  util.BuildConfig
	db *sqlx.DB
}

func NewIAPEnv(db *sqlx.DB, c util.BuildConfig) IAPEnv {
	return IAPEnv{
		c:  c,
		db: db,
	}
}

// BeginTx starts a transaction.
// NOTE: here the sandbox is different from the environment
// field send by apple. It only determines whether the
// sandbox db should be used and is determined by
// the CLI argument `-sandbox`.
// All messages from apple is save in production DBs.
func (env IAPEnv) BeginTx() (MembershipTx, error) {
	tx, err := env.db.Beginx()

	if err != nil {
		return MembershipTx{}, err
	}

	return MembershipTx{
		tx:      tx,
		sandbox: env.c.UseSandboxDB(),
	}, nil
}
