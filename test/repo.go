package test

import (
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

func (r *Repo) SaveAccount(a reader.Account) error {
	_, err := r.db.NamedExec(stmtInsertAccount, a)

	if err != nil {
		return err
	}

	return nil
}

func (r *Repo) MustSaveAccount(a reader.Account) {
	if err := r.SaveAccount(a); err != nil {
		panic(err)
	}
}

func (r *Repo) SaveMembership(m reader.Membership) error {
	m = m.Normalize()

	_, err := r.db.NamedExec(
		reader.StmtCreateMember,
		m)

	if err != nil {
		return err
	}

	return nil
}

func (r *Repo) MustSaveMembership(m reader.Membership) {

	err := r.SaveMembership(m)

	if err != nil {
		panic(err)
	}
}

func (r *Repo) SaveOrder(order subs.Order) error {

	var stmt = subs.StmtInsertOrder + `,
		confirmed_utc = :confirmed_utc,
		start_date = :start_date,
		end_date = :end_date`

	_, err := r.db.NamedExec(
		stmt,
		order)

	if err != nil {
		return err
	}

	return nil
}

func (r *Repo) MustSaveOrder(order subs.Order) {

	if err := r.SaveOrder(order); err != nil {
		panic(err)
	}
}

func (r *Repo) MustSaveRenewalOrders(orders []subs.Order) {
	for _, v := range orders {
		r.MustSaveOrder(v)
	}
}

// MustSaveProratedOrders inserts prorated orders
// to test ProratedOrdersUsed.
func (r *Repo) MustSaveProratedOrders(pos []subs.ProratedOrder) {

	for _, v := range pos {
		_, err := r.db.NamedExec(
			subs.StmtSaveProratedOrder,
			v)

		if err != nil {
			panic(err)
		}
	}
}
