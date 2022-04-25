package stripe

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/stripe/stripe-go/v72"
)

type InvoiceCollectionMethod struct {
	stripe.InvoiceCollectionMethod
}

func newInvoiceCollectionMethod(cm *stripe.InvoiceCollectionMethod) InvoiceCollectionMethod {
	if cm == nil {
		return InvoiceCollectionMethod{}
	}

	return InvoiceCollectionMethod{*cm}
}

func (cm InvoiceCollectionMethod) MarshalJSON() ([]byte, error) {
	return json.Marshal(cm.InvoiceCollectionMethod)
}

func (cm *InvoiceCollectionMethod) UnmarshalJSON(b []byte) error {
	if b == nil {
		*cm = InvoiceCollectionMethod{}
	}

	var v stripe.InvoiceCollectionMethod
	err := json.Unmarshal(b, &v)

	if err != nil {
		return err
	}

	*cm = InvoiceCollectionMethod{v}
	return nil
}

func (cm InvoiceCollectionMethod) Value() (driver.Value, error) {
	if cm.InvoiceCollectionMethod == "" {
		return nil, nil
	}

	return string(cm.InvoiceCollectionMethod), nil
}

func (cm *InvoiceCollectionMethod) Scan(src interface{}) error {
	if src == nil {
		*cm = InvoiceCollectionMethod{}
		return nil
	}

	switch s := src.(type) {
	case []byte:
		*cm = InvoiceCollectionMethod{stripe.InvoiceCollectionMethod(s)}
		return nil

	default:
		return errors.New("incompatible type to scan to PaymentIntent")
	}
}

type InvoiceStatus struct {
	stripe.InvoiceStatus
}

func (iv InvoiceStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(iv.InvoiceStatus)
}

func (iv *InvoiceStatus) UnmarshalJSON(b []byte) error {
	if b == nil {
		*iv = InvoiceStatus{}
	}

	var v stripe.InvoiceStatus
	err := json.Unmarshal(b, &v)

	if err != nil {
		return err
	}

	*iv = InvoiceStatus{v}
	return nil
}

func (iv InvoiceStatus) Value() (driver.Value, error) {
	if iv.InvoiceStatus == "" {
		return nil, nil
	}

	return string(iv.InvoiceStatus), nil
}

func (iv *InvoiceStatus) Scan(src interface{}) error {
	if src == nil {
		*iv = InvoiceStatus{}
		return nil
	}

	switch s := src.(type) {
	case []byte:
		*iv = InvoiceStatus{stripe.InvoiceStatus(s)}
		return nil

	default:
		return errors.New("incompatible type to scan to PaymentIntent")
	}
}

type SetupFutureUsage struct {
	stripe.PaymentIntentSetupFutureUsage
}

func (sf SetupFutureUsage) MarshalJSON() ([]byte, error) {
	return json.Marshal(sf.PaymentIntentSetupFutureUsage)
}

func (sf *SetupFutureUsage) UnmarshalJSON(b []byte) error {
	if b == nil {
		*sf = SetupFutureUsage{}
	}

	var v stripe.PaymentIntentSetupFutureUsage
	err := json.Unmarshal(b, &v)

	if err != nil {
		return err
	}

	*sf = SetupFutureUsage{v}
	return nil
}

func (sf SetupFutureUsage) Value() (driver.Value, error) {
	if sf.PaymentIntentSetupFutureUsage == "" {
		return nil, nil
	}

	return string(sf.PaymentIntentSetupFutureUsage), nil
}

func (sf *SetupFutureUsage) Scan(src interface{}) error {
	if src == nil {
		*sf = SetupFutureUsage{}
		return nil
	}

	switch s := src.(type) {
	case []byte:
		*sf = SetupFutureUsage{stripe.PaymentIntentSetupFutureUsage(s)}
		return nil

	default:
		return errors.New("incompatible type to scan to SetupFutureUsage")
	}
}

type SIStatus struct {
	stripe.SetupIntentStatus
}

func (ss SIStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(ss.SetupIntentStatus)
}

func (ss *SIStatus) UnmarshalJSON(b []byte) error {
	if b == nil {
		*ss = SIStatus{}
	}

	var v stripe.SetupIntentStatus
	err := json.Unmarshal(b, &v)

	if err != nil {
		return err
	}

	*ss = SIStatus{v}
	return nil
}

func (ss SIStatus) Value() (driver.Value, error) {
	if ss.SetupIntentStatus == "" {
		return nil, nil
	}

	return string(ss.SetupIntentStatus), nil
}

func (ss *SIStatus) Scan(src interface{}) error {
	if src == nil {
		*ss = SIStatus{}
		return nil
	}

	switch s := src.(type) {
	case []byte:
		*ss = SIStatus{stripe.SetupIntentStatus(s)}
		return nil

	default:
		return errors.New("incompatible type to scan to SIStatus")
	}
}

type SIUsage struct {
	stripe.SetupIntentUsage
}

func (su SIUsage) MarshalJSON() ([]byte, error) {
	return json.Marshal(su.SetupIntentUsage)
}

func (su *SIUsage) UnmarshalJSON(b []byte) error {
	if b == nil {
		*su = SIUsage{}
	}

	var v stripe.SetupIntentUsage
	err := json.Unmarshal(b, &v)

	if err != nil {
		return err
	}

	*su = SIUsage{v}
	return nil
}

func (su SIUsage) Value() (driver.Value, error) {
	if su.SetupIntentUsage == "" {
		return nil, nil
	}

	return string(su.SetupIntentUsage), nil
}

func (su *SIUsage) Scan(src interface{}) error {
	if src == nil {
		*su = SIUsage{}
		return nil
	}

	switch s := src.(type) {
	case []byte:
		*su = SIUsage{stripe.SetupIntentUsage(s)}
		return nil

	default:
		return errors.New("incompatible type to scan to SetupFutureUsage")
	}
}

type SICancelReason struct {
	stripe.SetupIntentCancellationReason
}

func (su SICancelReason) MarshalJSON() ([]byte, error) {
	return json.Marshal(su.SetupIntentCancellationReason)
}

func (su *SICancelReason) UnmarshalJSON(b []byte) error {
	if b == nil {
		*su = SICancelReason{}
	}

	var v stripe.SetupIntentCancellationReason
	err := json.Unmarshal(b, &v)

	if err != nil {
		return err
	}

	*su = SICancelReason{v}
	return nil
}

func (su SICancelReason) Value() (driver.Value, error) {
	if su.SetupIntentCancellationReason == "" {
		return nil, nil
	}

	return string(su.SetupIntentCancellationReason), nil
}

func (su *SICancelReason) Scan(src interface{}) error {
	if src == nil {
		*su = SICancelReason{}
		return nil
	}

	switch s := src.(type) {
	case []byte:
		*su = SICancelReason{stripe.SetupIntentCancellationReason(s)}
		return nil

	default:
		return errors.New("incompatible type to scan to SICancelReason")
	}
}
