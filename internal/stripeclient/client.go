package stripeclient

import (
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/stripe/stripe-go/v72"
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

// CreateEphemeralKey generate a key so that client could restrict customer API directly.
// Response:
// {
//  "id": "ephkey_1K9LFkBzTK0hABgJGBr4xtqx",
//  "object": "ephemeral_key",
//  "associated_objects": [
//    {
//      "type": "customer",
//      "id": "cus_KoyzEFNfmfdEpK"
//    }
//  ],
//  "created": 1640142752,
//  "expires": 1640146352,
//  "livemode": false,
//  "secret": "ek_test_YWNjdF8xRXBXM0VCelRLMGhBQmdKLHFBYmlRRGUyd1RUN0NMMU0yaHVkSG9PZlEyYjBvREY_00RG22HwXI"
// }
// Error response:
// {
//    "code": "resource_missing",
//    "doc_url": "https://stripe.com/docs/error-codes/resource-missing",
//    "status": 400,
//    "message": "No such customer: 'cus_IXp31Fk2jYJmU3'",
//    "param": "customer",
//    "request_id": "req_bXfWHybYVXFvWj",
//    "type": "invalid_request_error"
// }
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
