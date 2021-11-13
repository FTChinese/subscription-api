package price

import (
	"github.com/stripe/stripe-go/v72"
	"sync"
)

// stripePriceCache acts as a cache after retrieving Stripe prices.
type stripePriceCache struct {
	len     int
	prices  []StripePrice  // Use an array to store the prices.
	idIndex map[string]int // price id to its position in the array.
	mux     sync.Mutex     // The data is global. Lock it for concurrency.
}

func newStripePriceStore() *stripePriceCache {
	return &stripePriceCache{
		len:     0,
		prices:  make([]StripePrice, 0),
		idIndex: map[string]int{},
	}
}

func (store *stripePriceCache) Len() int {
	return store.len
}

// AddAll caches an array of stripe prices.
func (store *stripePriceCache) AddAll(sps []*stripe.Price) {
	store.mux.Lock()
	for _, sp := range sps {
		_ = store.upsert(sp)
	}
	store.mux.Unlock()
}

// Upsert inserts or update a price.
func (store *stripePriceCache) Upsert(sp *stripe.Price) error {
	store.mux.Lock()
	defer store.mux.Unlock()

	return store.upsert(sp)
}

func (store *stripePriceCache) upsert(sp *stripe.Price) error {

	p := NewStripePrice(sp)

	// If this price already cached, update it.
	// We perform this operation in case of any update on the Stripe side.
	// If a price is archived, we do not want to perform deletion
	// so that we don't need to move items one by one.
	index, ok := store.idIndex[p.ID]
	if ok {
		store.prices[index] = p
		return nil
	}

	// Now this price is not cached.
	// If stripe plan is not active, ignore it.
	if !sp.Active {
		return nil
	}

	// The price is not cached, append to the end.
	store.prices = append(store.prices, p)
	// Index it.
	store.idIndex[p.ID] = store.len
	store.len++
	return nil
}

// List returned a list all prices in the specified environment.
func (store *stripePriceCache) List(live bool) []StripePrice {
	var prices = make([]StripePrice, 0)

	for _, v := range store.prices {
		if v.LiveMode != live {
			continue
		}

		prices = append(prices, v)
	}

	return prices
}

var StripePriceCache = newStripePriceStore()
