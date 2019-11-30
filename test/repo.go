package test

import (
	"github.com/jmoiron/sqlx"
	"gitlab.com/ftchinese/subscription-api/models/paywall"
	"gitlab.com/ftchinese/subscription-api/models/plan"
	"gitlab.com/ftchinese/subscription-api/models/query"
	"gitlab.com/ftchinese/subscription-api/models/reader"
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
	db    *sqlx.DB
	query query.Builder
}

func NewRepo() Repo {
	return Repo{
		db:    DB,
		query: query.NewBuilder(false),
	}
}

func (r Repo) SaveAccount(a reader.Account) {
	_, err := r.db.NamedExec(stmtInsertAccount, a)
	if err != nil {
		panic(err)
	}
}

func (r Repo) SaveOrder(order paywall.Order) {

	var stmt = r.query.InsertOrder() + `,
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

func (r Repo) SaveMember(m paywall.Membership) {
	m.Normalize()

	_, err := r.db.NamedExec(
		r.query.InsertMember(),
		m)

	if err != nil {
		panic(err)
	}
}

func (r Repo) UpdateMember(m paywall.Membership) {
	m.Normalize()

	_, err := r.db.NamedExec(
		r.query.UpdateMember(m.MemberColumn()),
		m)

	if err != nil {
		panic(err)
	}
}

// SaveBalanceSources populate data to the proration table.
func (r Repo) SaveBalanceSources(p []paywall.ProrationSource) {
	for _, v := range p {
		_, err := r.db.NamedExec(
			r.query.InsertProration(),
			v)

		if err != nil {
			panic(err)
		}
	}
}

func (r Repo) SaveUpgradePlan(up paywall.UpgradePlan) {
	var data = struct {
		paywall.UpgradePlan
		plan.Plan
	}{
		UpgradePlan: up,
		Plan:        up.Plan,
	}

	_, err := r.db.NamedExec(
		r.query.InsertUpgradePlan(),
		data)

	if err != nil {
		panic(err)
	}
}
