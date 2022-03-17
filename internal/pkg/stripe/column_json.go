package stripe

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"github.com/stripe/stripe-go/v72"
)

type PINextActionJSON struct {
	*stripe.PaymentIntentNextAction
}

// Value implements Valuer interface by serializing an Invitation into
// JSON data.
func (na PINextActionJSON) Value() (driver.Value, error) {
	if na.PaymentIntentNextAction == nil {
		return nil, nil
	}

	b, err := json.Marshal(na.PaymentIntentNextAction)
	if err != nil {
		return nil, err
	}

	return string(b), nil
}

// Scan implements Valuer interface by deserializing an invitation field.
func (na *PINextActionJSON) Scan(src interface{}) error {
	if src == nil {
		*na = PINextActionJSON{}
		return nil
	}

	switch s := src.(type) {
	case []byte:
		var tmp stripe.PaymentIntentNextAction
		err := json.Unmarshal(s, &tmp)
		if err != nil {
			return err
		}
		*na = PINextActionJSON{&tmp}
		return nil

	default:
		return errors.New("incompatible type to scan to PaymentIntent")
	}
}

type SINextActionJSON struct {
	*stripe.SetupIntentNextAction
}

// Value implements Valuer interface by serializing an Invitation into
// JSON data.
func (na SINextActionJSON) Value() (driver.Value, error) {
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
func (na *SINextActionJSON) Scan(src interface{}) error {
	if src == nil {
		*na = SINextActionJSON{}
		return nil
	}

	switch s := src.(type) {
	case []byte:
		var tmp stripe.SetupIntentNextAction
		err := json.Unmarshal(s, &tmp)
		if err != nil {
			return err
		}
		*na = SINextActionJSON{&tmp}
		return nil

	default:
		return errors.New("incompatible type to scan to PaymentIntent")
	}
}
