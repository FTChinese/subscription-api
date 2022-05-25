package stripeclient

import (
	"github.com/FTChinese/subscription-api/pkg/price"
	"sync"
)

type SyncPrices struct {
	mux   sync.Mutex
	store map[string]price.StripePrice
}

func NewSyncPrices() *SyncPrices {
	return &SyncPrices{
		mux:   sync.Mutex{},
		store: make(map[string]price.StripePrice),
	}
}

func (s *SyncPrices) Add(p price.StripePrice) {
	s.mux.Lock()
	s.store[p.ID] = p
	s.mux.Unlock()
}

func (s *SyncPrices) Get() map[string]price.StripePrice {
	return s.store
}

func (s *SyncPrices) Map(keys []string) []price.StripePrice {
	var list = make([]price.StripePrice, 0)
	for _, k := range keys {
		p, ok := s.store[k]
		if !ok {
			continue
		}
		list = append(list, p)
	}

	return list
}
