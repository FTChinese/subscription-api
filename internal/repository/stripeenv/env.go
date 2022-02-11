package stripeenv

import (
	"github.com/FTChinese/subscription-api/internal/repository"
)

// Env extends PriceStore
type Env struct {
	PriceStore
	repository.StripeRepo
}

func NewEnv(repo repository.StripeRepo, store PriceStore) Env {
	return Env{
		PriceStore: store,
		StripeRepo: repo,
	}
}
