package stripe

import (
	"database/sql/driver"
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

type PISetupFutureUsage struct {
	stripe.PaymentIntentSetupFutureUsage
}

func (sf PISetupFutureUsage) Value() (driver.Value, error) {
	if sf.PaymentIntentSetupFutureUsage == "" {
		return nil, nil
	}

	return string(sf.PaymentIntentSetupFutureUsage), nil
}

func (sf *PISetupFutureUsage) Scan(src interface{}) error {
	if src == nil {
		*sf = PISetupFutureUsage{}
		return nil
	}

	switch s := src.(type) {
	case []byte:
		*sf = PISetupFutureUsage{stripe.PaymentIntentSetupFutureUsage(s)}
		return nil

	default:
		return errors.New("incompatible type to scan to PISetupFutureUsage")
	}
}

type PIStatus struct {
	stripe.PaymentIntentStatus
}

func (ps PIStatus) Value() (driver.Value, error) {
	if ps.PaymentIntentStatus == "" {
		return nil, nil
	}

	return string(ps.PaymentIntentStatus), nil
}

func (ps *PIStatus) Scan(src interface{}) error {
	if src == nil {
		*ps = PIStatus{}
		return nil
	}

	switch s := src.(type) {
	case []byte:
		*ps = PIStatus{stripe.PaymentIntentStatus(s)}
		return nil

	default:
		return errors.New("incompatible type to scan to PIStatus")
	}
}

type SIStatus struct {
	stripe.SetupIntentStatus
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
		return errors.New("incompatible type to scan to PISetupFutureUsage")
	}
}

type SICancelReason struct {
	stripe.SetupIntentCancellationReason
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
