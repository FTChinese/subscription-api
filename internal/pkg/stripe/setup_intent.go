package stripe

import (
	"github.com/FTChinese/subscription-api/lib/collection"
	"github.com/guregu/null"
	"github.com/stripe/stripe-go/v72"
)

type SetupIntent struct {
	IsFromStripe       bool                  `json:"-"`
	ID                 string                `json:"id" db:"id"`
	CancellationReason SICancelReason        `json:"cancellationReason" db:"cancellation_reason"`
	ClientSecret       string                `json:"clientSecret" db:"client_secret"`
	Created            int64                 `json:"-" db:"created"`
	CustomerID         string                `json:"customerId" db:"customer_id"`
	LiveMode           bool                  `json:"liveMode" db:"live_mode"`
	NextAction         SINextActionJSON      `json:"nextAction" db:"next_action"`
	PaymentMethodID    null.String           `json:"paymentMethodId" db:"payment_method_id"`
	PaymentMethodTypes collection.StringList `json:"-" db:"payment_method_types"`
	Status             SIStatus              `json:"status" db:"intent_status"`
	Usage              SIUsage               `json:"usage" db:"intent_usage"`
}

func NewSetupIntent(si *stripe.SetupIntent) SetupIntent {
	if si == nil {
		return SetupIntent{}
	}

	var cusID string
	if si.Customer != nil {
		cusID = si.Customer.ID
	}

	var pmID string
	if si.PaymentMethod != nil {
		pmID = si.PaymentMethod.ID
	}

	return SetupIntent{
		IsFromStripe:       true,
		ID:                 si.ID,
		CancellationReason: SICancelReason{si.CancellationReason},
		ClientSecret:       si.ClientSecret,
		Created:            si.Created,
		CustomerID:         cusID,
		LiveMode:           si.Livemode,
		NextAction: SINextActionJSON{
			si.NextAction,
		},
		PaymentMethodID:    null.NewString(pmID, pmID != ""),
		PaymentMethodTypes: si.PaymentMethodTypes,
		Status:             SIStatus{si.Status},
		Usage:              SIUsage{si.Usage},
	}
}
