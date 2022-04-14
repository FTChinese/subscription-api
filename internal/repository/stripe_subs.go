package repository

import "github.com/FTChinese/subscription-api/internal/pkg/stripe"

func (repo StripeRepo) UpsertSubs(s stripe.Subs, expanded bool) error {
	var stmt string
	if expanded {
		stmt = stripe.StmtUpsertSubsExpanded
	} else {
		stmt = stripe.StmtUpsertSubsNotExpanded
	}

	_, err := repo.dbs.Write.NamedExec(
		stmt,
		s,
	)
	if err != nil {
		return err
	}

	return nil
}

// RetrieveSubs retrieves the stripe subscription stored in our db.
func (repo StripeRepo) RetrieveSubs(id string) (stripe.Subs, error) {
	var s stripe.Subs
	err := repo.dbs.Read.Get(&s, stripe.StmtRetrieveSubs, id)
	if err != nil {
		return s, err
	}

	return s, nil
}
