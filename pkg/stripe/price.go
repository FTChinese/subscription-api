package stripe

import (
	"database/sql"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/guregu/null"
	"github.com/stripe/stripe-go/v72"
	"sync"
)

// NewPrice extract the essential data from a stripe price.
func NewPrice(p Plan, sp *stripe.Price) price.Price {
	return price.Price{
		ID:         sp.ID,
		Edition:    p.Edition,
		Active:     sp.Active,
		Currency:   price.Currency(sp.Currency),
		LiveMode:   sp.Livemode,
		Nickname:   null.NewString(sp.Nickname, sp.Nickname != ""),
		ProductID:  sp.Product.ID,
		Source:     price.SourceStripe,
		UnitAmount: float64(sp.UnitAmount / 100),
	}
}

// priceCache acts as a cache after retrieving Stripe prices.
type priceCache struct {
	len     int
	prices  []price.Price  // Use an array to store the prices.
	idIndex map[string]int // price id to its position in the array.
	mux     sync.Mutex     // The data is global. Lock it for currency.
}

func newPriceStore() *priceCache {
	return &priceCache{
		len:     0,
		prices:  make([]price.Price, 0),
		idIndex: map[string]int{},
	}
}

func (store *priceCache) Len() int {
	return store.len
}

// AddAll caches an array of stripe prices.
func (store *priceCache) AddAll(sps []*stripe.Price) {
	store.mux.Lock()
	for _, sp := range sps {
		_ = store.upsert(sp)
	}
	store.mux.Unlock()
}

// Upsert inserts or update a price.
func (store *priceCache) Upsert(sp *stripe.Price) error {
	store.mux.Lock()
	defer store.mux.Unlock()

	return store.upsert(sp)
}

func (store *priceCache) upsert(sp *stripe.Price) error {
	// Check if a stripe price id is allowed to present.
	plan, err := PlanStore.FindByID(sp.ID)
	if err != nil {
		return err
	}

	// If this price already cached, update it.
	index, ok := store.idIndex[sp.ID]
	if ok {
		store.prices[index] = NewPrice(plan, sp)
		return nil
	}

	// If stripe plan is not active and not added to the store, ignore it.
	if !sp.Active {
		return nil
	}

	// The price is not cached, append to the end.
	store.prices = append(store.prices, NewPrice(plan, sp))
	store.len++
	return nil
}

// Find searches a price by id in the specified environment.
func (store *priceCache) Find(id string, live bool) (price.Price, error) {
	index, ok := store.idIndex[id]
	if ok {
		return price.Price{}, sql.ErrNoRows
	}

	cachedPrice := store.prices[index]
	if cachedPrice.LiveMode != live {
		return price.Price{}, sql.ErrNoRows
	}

	return cachedPrice, nil
}

// List returned a list all prices in the specified environment.
func (store *priceCache) List(live bool) []price.Price {
	var prices = make([]price.Price, 0)

	for _, v := range store.prices {
		if v.LiveMode != live {
			continue
		}

		prices = append(prices, v)
	}

	return prices
}

var PriceCache = newPriceStore()
