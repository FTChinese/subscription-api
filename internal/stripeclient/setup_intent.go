package stripeclient

import (
	"github.com/FTChinese/subscription-api/internal/pkg/stripe"
	sdk "github.com/stripe/stripe-go/v72"
)

func (c Client) CreateSetupIntent(p stripe.SetupIntentParams) (*sdk.SetupIntent, error) {
	params := &sdk.SetupIntentParams{
		Customer: sdk.String(p.Customer),
	}

	return c.sc.SetupIntents.New(params)
}
