package stripe

import (
	"github.com/stripe/stripe-go/v72"
	"testing"
)

func TestInvoiceCollectionMethodJSON(t *testing.T) {
	cm := InvoiceCollectionMethod{stripe.InvoiceCollectionMethodSendInvoice}

	b, err := cm.MarshalJSON()

	if err != nil {
		t.Error(err)
	}

	t.Logf("%s", b)

	var parsed InvoiceCollectionMethod
	err = parsed.UnmarshalJSON(b)
	if err != nil {
		t.Error(err)
	}

	t.Logf("%s", parsed)
}

func TestInvoiceStatus(t *testing.T) {
	iv := InvoiceStatus{stripe.InvoiceStatusPaid}

	b, err := iv.MarshalJSON()

	if err != nil {
		t.Error(err)
	}

	t.Logf("%s", b)
}

func TestSetupFutureUsage(t *testing.T) {
	sf := SetupFutureUsage{stripe.PaymentIntentSetupFutureUsageOffSession}

	b, err := sf.MarshalJSON()

	if err != nil {
		t.Error(err)
	}

	t.Logf("%s", b)
}

func TestSetupIntentStatus(t *testing.T) {
	ss := SIStatus{stripe.SetupIntentStatusSucceeded}

	b, err := ss.MarshalJSON()

	if err != nil {
		t.Error(err)
	}

	t.Logf("%s", b)
}

func TestSetupIntentUsage(t *testing.T) {
	su := SIUsage{stripe.SetupIntentUsageOnSession}

	b, err := su.MarshalJSON()

	if err != nil {
		t.Error(err)
	}

	t.Logf("%s", b)
}

func TestCancelReason(t *testing.T) {
	su := SICancelReason{stripe.SetupIntentCancellationReasonFailedInvoice}

	b, err := su.MarshalJSON()

	if err != nil {
		t.Error(err)
	}

	t.Logf("%s", b)
}
