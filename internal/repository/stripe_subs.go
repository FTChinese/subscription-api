package repository

import "github.com/FTChinese/subscription-api/internal/pkg/stripe"

// UpsertSubs inserts or updates an existing subscription.
// Then payment_intent_id field only exists when expanded is true.
// To avoid setting it to null when we are updating an existing one,
// set expanded to false when expand parameters are not set.
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
