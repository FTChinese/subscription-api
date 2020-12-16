package stripe

import (
	"database/sql"
	"fmt"
	"github.com/FTChinese/subscription-api/pkg/product"
	"github.com/guregu/null"
	"github.com/stripe/stripe-go/v72"
	"sync"
)

type PricePreset struct {
	product.Edition
	Live bool
}

var presetPrices = map[string]PricePreset{
	"plan_FXZYLOEbcvj5Tx": {
		Edition: product.NewStdMonthEdition(),
		Live:    true,
	},
	"plan_FXZZUEDpToPlZK": {
		Edition: product.NewStdYearEdition(),
		Live:    true,
	},
	"plan_FXZbv1cDTsUKOg": {
		Edition: product.NewPremiumEdition(),
		Live:    true,
	},
	"plan_FOdgPTznDwHU4i": {
		Edition: product.NewStdMonthEdition(),
		Live:    false,
	},
	"plan_FOdfeaqzczp6Ag": {
		Edition: product.NewStdYearEdition(),
		Live:    false,
	},
	"plan_FOde0uAr0V4WmT": {
		Edition: product.NewPremiumEdition(),
		Live:    false,
	},
}

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

func NewPrice(preset PricePreset, price *stripe.Price) Price {
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

type priceStore struct {
	len     int
	prices  []Price
	idIndex map[string]int
	mux     sync.Mutex
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

func (store *priceStore) AddAll(sps []*stripe.Price) {
	store.mux.Lock()
	for _, sp := range sps {
		_ = store.upsert(sp)
	}
	store.mux.Unlock()
}

func (store *priceStore) Upsert(sp *stripe.Price) error {
	store.mux.Lock()
	defer store.mux.Unlock()

	return store.upsert(sp)
}

func (store *priceStore) upsert(sp *stripe.Price) error {
	preset, ok := presetPrices[sp.ID]
	if !ok {
		return fmt.Errorf("unknown stripe price %s", sp.ID)
	}

	index, ok := store.idIndex[sp.ID]
	if ok {
		store.prices[index] = NewPrice(preset, sp)
		return nil
	}

	// If stripe plan is not active and not added to the store, ignore it.
	if !sp.Active {
		return nil
	}

	store.prices = append(store.prices, NewPrice(preset, sp))
	store.len++
	return nil
}

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
