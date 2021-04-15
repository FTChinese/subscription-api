package striperepo

import "github.com/FTChinese/subscription-api/pkg/stripe"

// SaveSubsError saves any error in stripe response.
func (env Env) SaveSubsError(e stripe.APIError) error {
	_, err := env.DBs.Write.NamedExec(stripe.StmtSaveAPIError, e)

	if err != nil {
		return err
	}

	return nil
}

// UpsertSubs saves a reduced version os stripe subscription.
func (env Env) UpsertSubs(s stripe.Subs) error {
	_, err := env.DBs.Write.NamedExec(stripe.StmtInsertSubs, s)
	if err != nil {
		return err
	}

	return nil
}

// RetrieveSubs retrieves the stripe subscription stored in our db.
func (env Env) RetrieveSubs(id string) (stripe.Subs, error) {
	var s stripe.Subs
	err := env.DBs.Read.Get(&s, stripe.StmtRetrieveSubs, id)
	if err != nil {
		return s, err
	}

	return s, nil
}

// SubsExists checks whether we stored a stripe subscription.
func (env Env) SubsExists(id string) (bool, error) {
	var ok bool
	err := env.DBs.Read.Get(&ok, stripe.StmtSubsExists, id)
	if err != nil {
		return false, err
	}

	return ok, nil
}
