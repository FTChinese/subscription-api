package test

import (
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/jmoiron/sqlx"
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

func (r *Repo) MustSaveMembership(m subs.Membership) {

	m.Normalize()

	_, err := r.db.NamedExec(
		subs.StmtCreateMember(config.SubsDBProd),
		m)

	if err != nil {
		panic(err)
	}
}

func (r *Repo) MustSaveOrder(order subs.Order) {

	var stmt = subs.StmtCreateOrder(config.SubsDBProd) + `,
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

func (r *Repo) MustSaveRenewalOrders(orders []subs.Order) {
	for _, v := range orders {
		r.MustSaveOrder(v)
	}
}

// SaveProratedOrders inserts prorated orders
// to test ProratedOrdersUsed.
func (r *Repo) SaveProratedOrders(upgrade subs.UpgradeSchema) {

	for _, v := range upgrade.Sources {
		_, err := r.db.NamedExec(
			subs.StmtSaveProratedOrder(config.SubsDBProd),
			v)

		if err != nil {
			panic(err)
		}
	}
}
