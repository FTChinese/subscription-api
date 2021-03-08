package invoice

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/pkg/addon"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/guregu/null"
	"time"
)

// Invoice is the result of order successfully paid.
// For one-time purchase upgrading or switching from one-time purchase to Stripe, an invoice of
type Invoice struct {
	ID         string `json:"id" db:"id"`
	CompoundID string `json:"compoundId" db:"compound_id"`
	price.Edition
	dt.YearMonthDay
	AddOnSource   addon.Source   `json:"addOnSource" db:"addon_source"` // Only exists when OrderKind is AddOn.
	AppleTxID     null.String    `json:"appleTxId" db:"apple_tx_id"`    // For carry-over by switching to Apple, the apple original transaction id.
	OrderID       null.String    `json:"orderId" db:"order_id"`         // Which order created this invoice. For upgrading carry-over, it indicates which order caused the the carry-over.         // Carry over for stripe does not have order id.
	OrderKind     enum.OrderKind `json:"orderKind" db:"order_kind"`     // Always non-null
	PaidAmount    float64        `json:"paidAmount" db:"paid_amount"`
	PaymentMethod enum.PayMethod `json:"payMethod" db:"payment_method"`
	PriceID       null.String    `json:"priceId" db:"price_id"`
	StripeSubsID  null.String    `json:"stripeSubsId" db:"stripe_subs_id"` // For carry-over by switching to Stripe, the stripe subscription id.
	CreatedUTC    chrono.Time    `json:"createdUtc" db:"created_utc"`
	ConsumedUTC   chrono.Time    `json:"consumedUtc" db:"consumed_utc"`
	dt.DateTimePeriod
	CarriedOverUtc chrono.Time `json:"carriedOver" db:"carried_over_utc"` // In case user has carry-over for upgrading or switching stripe, add a timestamp to original invoice.
}

func (i Invoice) SetPeriod(start time.Time) Invoice {

	if start.IsZero() {
		return i
	}

	period := dt.NewTimeRange(start).
		WithDate(i.YearMonthDay).
		ToDateTimePeriod()

	i.ConsumedUTC = chrono.TimeNow()
	i.DateTimePeriod = period

	return i
}

func (i Invoice) IsConsumed() bool {
	return !i.ConsumedUTC.IsZero()
}

func (i Invoice) IsAddOn() bool {
	return i.OrderKind == enum.OrderKindAddOn
}

func (i Invoice) IsZero() bool {
	return i.ID == ""
}

// WithOrderID set the order id field for carried-over invoice caused by upgrading.
func (i Invoice) WithOrderID(id string) Invoice {
	i.OrderID = null.StringFrom(id)
	return i
}

func (i Invoice) WithStripeSubsID(id string) Invoice {
	i.StripeSubsID = null.StringFrom(id)
	return i
}

func (i Invoice) WithAppleTxID(id string) Invoice {
	i.AppleTxID = null.StringFrom(id)
	return i
}
