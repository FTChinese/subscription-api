package stripeclient

import (
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/client"
	"go.uber.org/zap"
)

type Client struct {
	live   bool // Deprecated
	sc     *client.API
	logger *zap.Logger
}

func New(live bool, logger *zap.Logger) Client {

	key := config.MustStripeAPIKey().Pick(live)

	return Client{
		live:   live,
		sc:     client.New(key, nil),
		logger: logger,
	}
}

// Deprecated
func (c Client) Live() bool {
	return c.live
}

// CreateEphemeralKey generate a key so that client could restrict customer API directly.
func (c Client) CreateEphemeralKey(cusID, version string) ([]byte, error) {
	params := &stripe.EphemeralKeyParams{
		Customer:      stripe.String(cusID),
		StripeVersion: stripe.String(version),
	}

	key, err := c.sc.EphemeralKeys.New(params)
	if err != nil {
		return nil, err
	}

	return key.RawJSON, nil
}
