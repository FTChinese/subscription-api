package shared

import (
	"github.com/FTChinese/subscription-api/internal/repository/stripeclient"
	"github.com/FTChinese/subscription-api/pkg/stripe"
)

type StripeBaseRepo struct {
	Client stripeclient.Client
	Live   bool // Deprecated
	Cache  *stripe.PriceCache
}

func (repo StripeBaseRepo) ListPrices(bustCache bool) ([]stripe.Price, error) {
	if !bustCache && repo.Cache.Len() != 0 {
		return repo.Cache.
			List(repo.Live), nil
	}

	sp, err := repo.Client.ListPrices()
	if err != nil {
		return nil, err
	}

	repo.Cache.AddAll(sp)

	return repo.Cache.
		List(repo.Live), nil
}

func (repo StripeBaseRepo) LoadPrice(id string, bustCache bool) (stripe.Price, error) {
	if !bustCache {
		p, ok := repo.Cache.Find(id)
		if ok {
			return p, nil
		}
	}

	sp, err := repo.Client.RetrievePrice(id)
	if err != nil {
		return stripe.Price{}, err
	}

	p := stripe.NewPrice(sp)

	repo.Cache.Upsert(p)

	return p, nil
}
