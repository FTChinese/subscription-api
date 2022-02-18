package stripe

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/lib/collection"
	"github.com/FTChinese/subscription-api/lib/validator"
	"github.com/guregu/null"
	"github.com/stripe/stripe-go/v72"
)

type SetupIntentParams struct {
	Customer string `json:"customer"`
}

func (p SetupIntentParams) Validate() *render.ValidationError {
	return validator.New("customer").Required().Validate(p.Customer)
}

type SetupIntentNextActionJSON struct {
	*stripe.SetupIntentNextAction
}

// Value implements Valuer interface by serializing an Invitation into
// JSON data.
func (na SetupIntentNextActionJSON) Value() (driver.Value, error) {
	if na.SetupIntentNextAction == nil {
		return nil, nil
	}

	b, err := json.Marshal(na.SetupIntentNextAction)
	if err != nil {
		return nil, err
	}

	return string(b), nil
}

// Scan implements Valuer interface by deserializing an invitation field.
func (na *SetupIntentNextActionJSON) Scan(src interface{}) error {
	if src == nil {
		*na = SetupIntentNextActionJSON{}
		return nil
	}

	switch s := src.(type) {
	case []byte:
		var tmp stripe.SetupIntentNextAction
		err := json.Unmarshal(s, &tmp)
		if err != nil {
			return err
		}
		*na = SetupIntentNextActionJSON{&tmp}
		return nil

	default:
		return errors.New("incompatible type to scan to PaymentIntent")
	}
}

type SetupIntent struct {
	ID                 string                               `json:"id" db:"id"`
	CancellationReason stripe.SetupIntentCancellationReason `json:"cancellationReason" db:"cancellation_reason"`
	ClientSecret       string                               `json:"clientSecret" db:"client_secret"`
	Created            int64                                `json:"-" db:"created"`
	CustomerID         string                               `json:"customerId" db:"customer_id"`
	LiveMode           bool                                 `json:"liveMode" db:"live_mode"`
	NextAction         SetupIntentNextActionJSON            `json:"next_action" db:"next_action"`
	PaymentMethodID    null.String                          `json:"paymentMethodId" db:"payment_method_id"`
	PaymentMethodTypes collection.StringList                `json:"-" db:"payment_method_types"`
	Status             stripe.SetupIntentStatus             `json:"status" db:"intent_status"`
	Usage              stripe.SetupIntentUsage              `json:"usage" db:"intent_usage"`
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
		ID:                 si.ID,
		CancellationReason: si.CancellationReason,
		ClientSecret:       si.ClientSecret,
		Created:            si.Created,
		CustomerID:         cusID,
		LiveMode:           si.Livemode,
		NextAction: SetupIntentNextActionJSON{
			si.NextAction,
		},
		PaymentMethodID:    null.NewString(pmID, pmID != ""),
		PaymentMethodTypes: si.PaymentMethodTypes,
		Status:             si.Status,
		Usage:              si.Usage,
	}
}
