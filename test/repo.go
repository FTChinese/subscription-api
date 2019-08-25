package test

import (
	"github.com/jmoiron/sqlx"
	"gitlab.com/ftchinese/subscription-api/models/paywall"
	"gitlab.com/ftchinese/subscription-api/models/query"
)

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
