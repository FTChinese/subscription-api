package test

import (
	"github.com/jmoiron/sqlx"
	"gitlab.com/ftchinese/subscription-api/models/plan"
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

func NewRepo() Repo {
	return Repo{
		db: DB,
	}
}

func (r Repo) SaveAccount(a reader.Account) {
	_, err := r.db.NamedExec(stmtInsertAccount, a)
	if err != nil {
		panic(err)
	}
}

func (r Repo) SaveOrder(order subscription.Order) {

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

func (r Repo) SaveMember(m subscription.Membership) {
	m.Normalize()

	_, err := r.db.NamedExec(
		query.BuildInsertMembership(false),
		m)

	if err != nil {
		panic(err)
	}
}

func (r Repo) UpdateMember(m subscription.Membership) {
	m.Normalize()

	_, err := r.db.NamedExec(
		query.BuildUpdateMembership(false),
		m)

	if err != nil {
		panic(err)
	}
}

// SaveBalanceSources populate data to the proration table.
func (r Repo) SaveBalanceSources(p []plan.ProrationSource) {
	for _, v := range p {
		_, err := r.db.NamedExec(
			query.BuildInsertProration(false),
			v)

		if err != nil {
			panic(err)
		}
	}
}

func (r Repo) SaveUpgradePlan(up plan.UpgradePlan) {
	var data = struct {
		plan.UpgradePlan
		plan.Plan
	}{
		UpgradePlan: up,
		Plan:        up.Plan,
	}

	_, err := r.db.NamedExec(
		query.BuildInsertUpgradePlan(false),
		data)

	if err != nil {
		panic(err)
	}
}
