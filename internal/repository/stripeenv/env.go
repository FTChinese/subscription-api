package stripeenv

import (
	"github.com/FTChinese/subscription-api/internal/repository"
	"github.com/FTChinese/subscription-api/internal/stripeclient"
)

type Env struct {
	Client stripeclient.Client
	repository.StripeRepo
}

func New(client stripeclient.Client, repo repository.StripeRepo) Env {
	return Env{
		Client:     client,
		StripeRepo: repo,
	}
}
