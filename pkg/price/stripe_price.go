package price

import (
	"database/sql"
	"github.com/guregu/null"
	"github.com/stripe/stripe-go/v72"
	"sync"
)

// NewStripe extract the essential data from a stripe price.
func NewStripe(p StripeEdition, sp *stripe.Price) Price {
	return Price{
		ID:         sp.ID,
		Edition:    p.Edition,
		Active:     sp.Active,
		Currency:   Currency(sp.Currency),
		LiveMode:   sp.Livemode,
		Nickname:   null.NewString(sp.Nickname, sp.Nickname != ""),
		ProductID:  sp.Product.ID,
		Source:     SourceStripe,
		UnitAmount: float64(sp.UnitAmount) / 100,
	}
}

// stripePriceCache acts as a cache after retrieving Stripe prices.
type stripePriceCache struct {
	len     int
	prices  []Price        // Use an array to store the prices.
	idIndex map[string]int // price id to its position in the array.
	mux     sync.Mutex     // The data is global. Lock it for currency.
}

func newStripePriceStore() *stripePriceCache {
	return &stripePriceCache{
		len:     0,
		prices:  make([]Price, 0),
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
	// Check if a stripe price id is allowed to present.
	plan, err := StripeEditions.FindByID(sp.ID)
	if err != nil {
		return err
	}

	// If this price already cached, update it.
	index, ok := store.idIndex[sp.ID]
	if ok {
		store.prices[index] = NewStripe(plan, sp)
		return nil
	}

	// If stripe plan is not active and not added to the store, ignore it.
	if !sp.Active {
		return nil
	}

	// The price is not cached, append to the end.
	store.prices = append(store.prices, NewStripe(plan, sp))
	store.len++
	return nil
}

// Find searches a price by id in the specified environment.
func (store *stripePriceCache) Find(id string, live bool) (Price, error) {
	index, ok := store.idIndex[id]
	if ok {
		return Price{}, sql.ErrNoRows
	}

	cachedPrice := store.prices[index]
	if cachedPrice.LiveMode != live {
		return Price{}, sql.ErrNoRows
	}

	return cachedPrice, nil
}

// List returned a list all prices in the specified environment.
func (store *stripePriceCache) List(live bool) []Price {
	var prices = make([]Price, 0)

	for _, v := range store.prices {
		if v.LiveMode != live {
			continue
		}

		prices = append(prices, v)
	}

	return prices
}

var StripePriceCache = newStripePriceStore()
