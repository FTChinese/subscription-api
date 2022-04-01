package stripeclient

import (
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/stripe/stripe-go/v72/client"
	"go.uber.org/zap"
)

type Client struct {
	sc     *client.API
	logger *zap.Logger
}

func New(live bool, logger *zap.Logger) Client {

	key := config.MustStripeAPIKey().Pick(live)

	return Client{
		sc:     client.New(key, nil),
		logger: logger,
	}
}

func (c Client) Get() *client.API {
	return c.sc
}
