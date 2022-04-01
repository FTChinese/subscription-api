package stripeclient

import "github.com/stripe/stripe-go/v72"

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

type asyncEphemeralKey struct {
	value *stripe.EphemeralKey
	error error
}

func (c Client) asyncCreateEphemeralKey(cusID, version string) <-chan asyncEphemeralKey {
	ch := make(chan asyncEphemeralKey)

	params := &stripe.EphemeralKeyParams{
		Customer:      stripe.String(cusID),
		StripeVersion: stripe.String(version),
	}

	go func() {
		defer close(ch)

		key, err := c.sc.EphemeralKeys.New(params)
		ch <- asyncEphemeralKey{
			value: key,
			error: err,
		}
	}()

	return ch
}
