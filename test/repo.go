package test

import (
	"github.com/jmoiron/sqlx"
	"gitlab.com/ftchinese/subscription-api/models/reader"
	"gitlab.com/ftchinese/subscription-api/models/subscription"
	"gitlab.com/ftchinese/subscription-api/repository/query"
)

const stmtInsertAccount = `
INSERT INTO cmstmp01.userinfo
SET user_id = :ftc_id,
	wx_union_id = :union_id,
	stripe_customer_id = :stripe_id,
	user_name = :user_name,
	email = :email,
	password = '12345678'`

type Repo struct {
	db *sqlx.DB
}

func NewRepo() *Repo {
	return &Repo{
		db: DB,
	}
}

func (r *Repo) MustCreateAccount(a reader.Account) {
	_, err := r.db.NamedExec(stmtInsertAccount, a)

	if err != nil {
		panic(err)
	}
}

func (r *Repo) MustSaveMembership(m subscription.Membership) {

	m.Normalize()

	_, err := r.db.NamedExec(
		query.BuildInsertMembership(false),
		m)

	if err != nil {
		panic(err)
	}
}

func (r *Repo) MustSaveOrder(order subscription.Order) {

	var stmt = query.BuildInsertOrder(false) + `,
		confirmed_utc = :confirmed_at,
		start_date = :start_date,
		end_date = :end_date`

	_, err := r.db.NamedExec(
		stmt,
		order)

	if err != nil {
		panic(err)
	}
}

func (r *Repo) MustSaveRenewalOrders(orders []subscription.Order) {
	for _, v := range orders {
		r.MustSaveOrder(v)
	}
}

// SaveProratedOrders inserts prorated orders
// to test ProratedOrdersUsed.
func (r *Repo) SaveProratedOrders(upgrade subscription.UpgradeSchema) {

	for _, v := range upgrade.Sources {
		_, err := r.db.NamedExec(
			query.BuildInsertProration(false),
			v)

		if err != nil {
			panic(err)
		}
	}
}
