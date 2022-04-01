package stripeclient

import (
	"github.com/FTChinese/subscription-api/internal/pkg/stripe"
	sdk "github.com/stripe/stripe-go/v72"
)

func (c Client) CreateSetupIntent(p stripe.CustomerParams) (*sdk.SetupIntent, error) {
	params := &sdk.SetupIntentParams{
		Customer: sdk.String(p.Customer),
	}

	return c.sc.SetupIntents.New(params)
}

type asyncSetupIntent struct {
	value *sdk.SetupIntent
	error error
}

func (c Client) asyncCreateSetupIntent(cusID string) <-chan asyncSetupIntent {
	ch := make(chan asyncSetupIntent)

	params := &sdk.SetupIntentParams{
		Customer: sdk.String(cusID),
	}

	go func() {
		si, err := c.sc.SetupIntents.New(params)
		ch <- asyncSetupIntent{
			value: si,
			error: err,
		}
	}()

	return ch
}

func (c Client) SetupWithEphemeral(cusID, version string) (*sdk.SetupIntent, *sdk.EphemeralKey, error) {
	setupCh, keyCh := c.asyncCreateSetupIntent(cusID), c.asyncCreateEphemeralKey(cusID, version)

	setupResult, keyResult := <-setupCh, <-keyCh
	if setupResult.error != nil {
		return nil, nil, setupResult.error
	}
	if keyResult.error != nil {
		return nil, nil, keyResult.error
	}

	return setupResult.value, keyResult.value, nil
}

func (c Client) FetchSetupIntent(id string, expandPaymentMethod bool) (*sdk.SetupIntent, error) {
	var params = &sdk.SetupIntentParams{}
	if expandPaymentMethod {
		params.AddExpand("payment_method")
	}

	return c.sc.SetupIntents.Get(id, params)
}
