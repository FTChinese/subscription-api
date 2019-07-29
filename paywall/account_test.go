package paywall

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"github.com/icrowley/fake"
	"github.com/stripe/stripe-go"
	"testing"
)

func newFtcUser() Account {
	return Account{
		FtcID:    uuid.New().String(),
		UnionID:  null.String{},
		StripeID: null.String{},
		Email:    fake.EmailAddress(),
		UserName: null.StringFrom(fake.UserName()),
	}
}

func TestFtcUser_StripeSubParcel(t *testing.T) {
	var s stripe.Subscription

	if err := json.Unmarshal([]byte(subDataExpanded), &s); err != nil {
		t.Error(err)
	}

	ss := NewStripeSub(&s)

	parcel, err := newFtcUser().StripeSubParcel(ss)
	if err != nil {
		t.Error(err)
	}

	t.Logf("%+v", parcel)
}

func TestFtcUser_StripeInvoiceParcel(t *testing.T) {
	var i stripe.Invoice

	if err := json.Unmarshal([]byte(invoiceData), &i); err != nil {
		t.Error(err)
	}

	parcel, err := newFtcUser().StripeInvoiceParcel(EmailedInvoice{&i})
	if err != nil {
		t.Error(err)
	}

	t.Logf("%+v", parcel)
}
