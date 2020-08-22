package iaprepo

import "github.com/FTChinese/subscription-api/pkg/apple"

// CreateSubscription saves an Subscription instance
// built from the latest transaction.
func (env Env) CreateSubscription(s apple.Subscription) error {
	_, err := env.db.NamedExec(apple.StmtUpsertSubs, s)

	if err != nil {
		logger.WithField("trace", "Env.CreateSubscription").Error(err)
		return err
	}

	return nil
}

func (env Env) LoadSubscription(originalID string) (apple.Subscription, error) {
	var s apple.Subscription
	err := env.db.Get(&s, apple.StmtLoadSubs, originalID)

	if err != nil {
		return apple.Subscription{}, err
	}

	return s, nil
}
