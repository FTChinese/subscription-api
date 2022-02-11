package stripeenv

import (
	"github.com/FTChinese/subscription-api/internal/pkg/stripe"
	"github.com/FTChinese/subscription-api/internal/stripeclient"
)

// PriceStore wraps shared features to manipulate stripe
// prices.
// It is meant to be extended whenever you need to access
// stripeclient instance.
type PriceStore struct {
	Client stripeclient.Client
	Cache  *stripe.PriceCache
}

func (store PriceStore) ListPrices(live bool, bustCache bool) ([]stripe.Price, error) {
	if !bustCache && store.Cache.Len() != 0 {
		return store.Cache.
			List(live), nil
	}

	sp, err := store.Client.ListPrices()
	if err != nil {
		return nil, err
	}

	store.Cache.AddAll(sp)

	return store.Cache.
		List(live), nil
}

func (store PriceStore) LoadPrice(id string, bustCache bool) (stripe.Price, error) {
	if !bustCache {
		p, ok := store.Cache.Find(id)
		if ok {
			return p, nil
		}
	}

	sp, err := store.Client.RetrievePrice(id)
	if err != nil {
		return stripe.Price{}, err
	}

	p := stripe.NewPrice(sp)

	store.Cache.Upsert(p)

	return p, nil
}

func (store PriceStore) LoadCheckoutItem(params stripe.SubsParams) (stripe.CheckoutItem, error) {
	p, err := store.LoadPrice(params.PriceID, false)
	if err != nil {
		return stripe.CheckoutItem{}, err
	}

	if params.IntroductoryPriceID.IsZero() {
		return stripe.CheckoutItem{
			Price:        p,
			Introductory: stripe.Price{},
		}, nil
	}

	introPrice, err := store.LoadPrice(params.IntroductoryPriceID.String, false)
	if err != nil {
		return stripe.CheckoutItem{}, err
	}

	return stripe.CheckoutItem{
		Price:        p,
		Introductory: introPrice,
	}, nil
}

func (store PriceStore) UpdatePriceMeta(id string, m map[string]string) (stripe.Price, error) {
	p, err := store.Client.SetPriceMeta(id, m)
	if err != nil {
		return stripe.Price{}, err
	}

	return stripe.NewPrice(p), nil
}
