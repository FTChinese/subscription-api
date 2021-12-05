package shared

import (
	"github.com/FTChinese/subscription-api/internal/pkg/stripe"
	"github.com/FTChinese/subscription-api/internal/repository/stripeclient"
)

type StripeBaseRepo struct {
	Client stripeclient.Client
	Cache  *stripe.PriceCache
}

func NewStripeCommon(client stripeclient.Client, c *stripe.PriceCache) StripeBaseRepo {
	return StripeBaseRepo{
		Client: client,
		Cache:  c,
	}
}

func (repo StripeBaseRepo) ListPrices(live bool, bustCache bool) ([]stripe.Price, error) {
	if !bustCache && repo.Cache.Len() != 0 {
		return repo.Cache.
			List(live), nil
	}

	sp, err := repo.Client.ListPrices()
	if err != nil {
		return nil, err
	}

	repo.Cache.AddAll(sp)

	return repo.Cache.
		List(live), nil
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

func (repo StripeBaseRepo) LoadCheckoutItem(params stripe.SubsParams) (stripe.CheckoutItem, error) {
	p, err := repo.LoadPrice(params.PriceID, false)
	if err != nil {
		return stripe.CheckoutItem{}, err
	}

	if params.IntroductoryPriceID.IsZero() {
		return stripe.CheckoutItem{
			Price:        p,
			Introductory: stripe.Price{},
		}, nil
	}

	introPrice, err := repo.LoadPrice(params.IntroductoryPriceID.String, false)
	if err != nil {
		return stripe.CheckoutItem{}, err
	}

	return stripe.CheckoutItem{
		Price:        p,
		Introductory: introPrice,
	}, nil
}
