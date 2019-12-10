package test

import (
	"github.com/jmoiron/sqlx"
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
	store *SubStore
	db    *sqlx.DB
}

func NewRepo(store *SubStore) *Repo {
	return &Repo{
		store: store,
		db:    DB,
	}
}

func (r *Repo) MustCreateAccount() {
	_, err := r.db.NamedExec(stmtInsertAccount, r.store.GetAccount())

	if err != nil {
		panic(err)
	}
}

func (r *Repo) MustSaveMembership() subscription.Membership {
	m := r.store.MustGetMembership()

	m.Normalize()

	_, err := r.db.NamedExec(
		query.BuildInsertMembership(false),
		m)

	if err != nil {
		panic(err)
	}

	return m
}

func (r *Repo) mustSaveOrder(order subscription.Order) {
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

func (r *Repo) MustCreateOrder() subscription.Order {

	o := r.store.MustCreateOrder()

	r.mustSaveOrder(o)

	return o
}

// MustRenewN prepares data to test FindBalanceSources
func (r *Repo) MustRenewN(n int) {
	orders := r.store.MustRenewN(n)

	for _, v := range orders {
		r.mustSaveOrder(v)
	}
}

// SaveProratedOrders inserts prorated orders
// to test ProratedOrdersUsed.
func (r *Repo) SaveProratedOrders(n int) subscription.UpgradeSchema {
	upgrade, _ := r.store.MustUpgrade(n)

	for _, v := range upgrade.Sources {
		_, err := r.db.NamedExec(
			query.BuildInsertProration(false),
			v)

		if err != nil {
			panic(err)
		}
	}

	return upgrade
}
