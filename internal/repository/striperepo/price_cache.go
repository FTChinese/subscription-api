package striperepo

import (
	ftcStripe "github.com/FTChinese/subscription-api/pkg/stripe"
	"github.com/stripe/stripe-go/v72"
	"sync"
)

// priceCache acts as a cache after retrieving Stripe prices.
type priceCache struct {
	len     int
	prices  []ftcStripe.Price // Use an array to store the prices.
	idIndex map[string]int    // price id to its position in the array.
	mux     sync.Mutex        // The data is global. Lock it for concurrency.
}

func newPriceCache() *priceCache {
	return &priceCache{
		len:     0,
		prices:  make([]ftcStripe.Price, 0),
		idIndex: map[string]int{},
	}
}

func (store *priceCache) Len() int {
	return store.len
}

// AddAll caches an array of stripe prices.
func (store *priceCache) AddAll(sps []*stripe.Price) {
	store.mux.Lock()
	defer store.mux.Unlock()

	for _, sp := range sps {
		_ = store.upsert(ftcStripe.NewPrice(sp))
	}
}

// Upsert inserts or update a price.
func (store *priceCache) Upsert(p ftcStripe.Price) int {
	store.mux.Lock()
	defer store.mux.Unlock()

	return store.upsert(p)
}

func (store *priceCache) upsert(p ftcStripe.Price) int {

	// If this price already cached, update it.
	// We perform this operation in case of any update on the Stripe side.
	// If a price is archived, we do not want to perform deletion
	// so that we don't need to move items one by one.
	index, ok := store.idIndex[p.ID]
	if ok {
		store.prices[index] = p
		return index
	}

	// Now this price is not cached.
	// If stripe plan is not active, ignore it.
	if !p.Active {
		return -1
	}

	// The price is not cached, append to the end.
	store.prices = append(store.prices, p)
	// Index it.
	store.idIndex[p.ID] = store.len
	store.len++
	return store.len - 1
}

// List returned a list all prices in the specified environment.
func (store *priceCache) List(live bool) []ftcStripe.Price {
	var prices = make([]ftcStripe.Price, 0)

	for _, v := range store.prices {
		if v.LiveMode != live {
			continue
		}

		prices = append(prices, v)
	}

	return prices
}

// Find tries to find a Price by id.
func (store *priceCache) Find(id string) (ftcStripe.Price, bool) {
	i, ok := store.idIndex[id]
	if !ok {
		return ftcStripe.Price{}, false
	}

	return store.prices[i], true
}

var PriceCache = newPriceCache()
