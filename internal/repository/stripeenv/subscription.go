package stripeenv

import "github.com/FTChinese/subscription-api/internal/pkg/stripe"

func (env Env) LoadOrFetchSubs(id string, refresh bool) (stripe.Subs, error) {
	if !refresh {
		subs, err := env.RetrieveSubs(id)
		if err == nil {
			return subs, nil
		}
	}

	rawSubs, err := env.Client.FetchSubs(id, false)
	if err != nil {
		return stripe.Subs{}, err
	}

	return stripe.NewSubs("", rawSubs), nil
}
