package stripe

import (
	"database/sql"
	"fmt"
	"github.com/FTChinese/subscription-api/pkg/product"
	"github.com/guregu/null"
	"github.com/stripe/stripe-go/v72"
	"sync"
)

// PriceEdition defines the product edition used in a Stripe environment.
type PriceEdition struct {
	product.Edition
	Live bool
}

// The map between Stripe id to edition.
var presetPrices = map[string]PriceEdition{
	"plan_FXZYLOEbcvj5Tx": {
		Edition: product.StdMonthEdition,
		Live:    true,
	},
	"plan_FXZZUEDpToPlZK": {
		Edition: product.StdYearEdition,
		Live:    true,
	},
	"plan_FXZbv1cDTsUKOg": {
		Edition: product.PremiumEdition,
		Live:    true,
	},
	"plan_FOdgPTznDwHU4i": {
		Edition: product.StdMonthEdition,
		Live:    false,
	},
	"plan_FOdfeaqzczp6Ag": {
		Edition: product.StdYearEdition,
		Live:    false,
	},
	"plan_FOde0uAr0V4WmT": {
		Edition: product.PremiumEdition,
		Live:    false,
	},
}

func FindPriceEdition(priceID string) (product.Edition, error) {
	pe, ok := presetPrices[priceID]
	if !ok {
		return product.Edition{}, fmt.Errorf("stripe price with id %s is not found", priceID)
	}

	return pe.Edition, nil
}

// Price is a condensed version of Stripe's price object.
type Price struct {
	ID string `json:"id"`
	product.Edition
	Active     bool        `json:"active"`
	Currency   string      `json:"currency"`
	LiveMode   bool        `json:"liveMode"`
	Nickname   null.String `json:"nickname"`
	ProductID  string      `json:"productId"`
	UnitAmount int64       `json:"unitAmount"`
	Created    int64       `json:"created"`
}

// NewPrice extract the essential data from a stripe price.
func NewPrice(preset PriceEdition, price *stripe.Price) Price {
	return Price{
		ID:         price.ID,
		Edition:    preset.Edition,
		Active:     price.Active,
		Currency:   string(price.Currency),
		LiveMode:   price.Livemode,
		Nickname:   null.NewString(price.Nickname, price.Nickname != ""),
		ProductID:  price.Product.ID,
		UnitAmount: price.UnitAmount,
		Created:    price.Created,
	}
}

// priceStore acts as a cache after retrieving Stripe prices.
type priceStore struct {
	len     int
	prices  []Price        // Use an array to store the prices.
	idIndex map[string]int // price id to its position in the array.
	mux     sync.Mutex     // The data is global. Lock it for currency.
}

func newPriceStore() *priceStore {
	return &priceStore{
		len:     0,
		prices:  make([]Price, 0),
		idIndex: map[string]int{},
	}
}

func (store *priceStore) Len() int {
	return store.len
}

// AddAll caches an array of stripe prices.
func (store *priceStore) AddAll(sps []*stripe.Price) {
	store.mux.Lock()
	for _, sp := range sps {
		_ = store.upsert(sp)
	}
	store.mux.Unlock()
}

// Upsert inserts or update a price.
func (store *priceStore) Upsert(sp *stripe.Price) error {
	store.mux.Lock()
	defer store.mux.Unlock()

	return store.upsert(sp)
}

func (store *priceStore) upsert(sp *stripe.Price) error {
	// Check if a stripe price id is allowed to present.
	preset, ok := presetPrices[sp.ID]
	if !ok {
		return fmt.Errorf("unknown stripe price %s", sp.ID)
	}

	// If this price already cached, update it.
	index, ok := store.idIndex[sp.ID]
	if ok {
		store.prices[index] = NewPrice(preset, sp)
		return nil
	}

	// If stripe plan is not active and not added to the store, ignore it.
	if !sp.Active {
		return nil
	}

	// The price is not cached, append to the end.
	store.prices = append(store.prices, NewPrice(preset, sp))
	store.len++
	return nil
}

// Find searches a price by id in the specified environment.
func (store *priceStore) Find(id string, live bool) (Price, error) {
	index, ok := store.idIndex[id]
	if ok {
		return Price{}, sql.ErrNoRows
	}

	price := store.prices[index]
	if price.LiveMode != live {
		return Price{}, sql.ErrNoRows
	}

	return price, nil
}

// List returned a list all prices in the specified environment.
func (store *priceStore) List(live bool) []Price {
	var prices = make([]Price, 0)

	for _, v := range store.prices {
		if v.LiveMode != live {
			continue
		}

		prices = append(prices, v)
	}

	return prices
}

var PriceStore = newPriceStore()
