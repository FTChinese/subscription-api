package striperepo

import "github.com/FTChinese/subscription-api/pkg/stripe"

// SaveSubsError saves any error in stripe response.
func (env Env) SaveSubsError(e stripe.APIError) error {
	_, err := env.db.NamedExec(stripe.StmtSaveAPIError, e)

	if err != nil {
		return err
	}

	return nil
}

func (env Env) UpsertSubs(s stripe.Subs) error {
	_, err := env.db.NamedExec(stripe.StmtInsertSubs, s)
	if err != nil {
		return err
	}

	return nil
}

func (env Env) RetrieveSubs(id string) (stripe.Subs, error) {
	var s stripe.Subs
	err := env.db.Get(&s, stripe.StmtRetrieveSubs, id)
	if err != nil {
		return s, err
	}

	return s, nil
}

func (env Env) SubsExists(id string) (bool, error) {
	var ok bool
	err := env.db.Get(&ok, stripe.StmtSubsExists, id)
	if err != nil {
		return false, err
	}

	return ok, nil
}
