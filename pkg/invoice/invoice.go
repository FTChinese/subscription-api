package invoice

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/pkg/addon"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/guregu/null"
	"time"
)

// Invoice is the result of a successfully paid order.
//
// Currently invoices comes from two sources:
// * an order user paid;
// * remaining valid time of current subscription when switching to another subscription plan.
//
// When AddOnSource is addon.SourceCarryOver, an invoice is
// is generated by either of:
// * User selected to purchase an add-on;
// * An one-time-purchase standard edition changed to premium,
// the remaining period of current standard subscription is frozen
// into a Invoice;
// * An one-time-purchase membership, regardless of current edition,
// decides to select subscription model, either via IAP or stripe,
// and the remaining period of current subscription is frozen into a Invoice.
type Invoice struct {
	ID         string `json:"id" db:"id"`
	CompoundID string `json:"compoundId" db:"compound_id"`
	price.Edition
	dt.YearMonthDay
	AddOnSource   addon.Source   `json:"addOnSource" db:"addon_source"` // Only exists when OrderKind is AddOn.
	AppleTxID     null.String    `json:"appleTxId" db:"apple_tx_id"`    // The apple original transaction id which caused this invoice to be generated as a carry_over kind.
	OrderID       null.String    `json:"orderId" db:"order_id"`         // Which order created this invoice. For upgrading carry-over, it indicates which order caused the the carry-over.         // Carry over for stripe does not have order id.
	OrderKind     enum.OrderKind `json:"orderKind" db:"order_kind"`     // Which kind of order created this invoice. For addon, there's actually no original order.
	PaidAmount    float64        `json:"paidAmount" db:"paid_amount"`
	PaymentMethod enum.PayMethod `json:"payMethod" db:"payment_method"`
	PriceID       null.String    `json:"priceId" db:"price_id"`
	StripeSubsID  null.String    `json:"stripeSubsId" db:"stripe_subs_id"` // For carry-over by switching to Stripe, the stripe subscription id.
	CreatedUTC    chrono.Time    `json:"createdUtc" db:"created_utc"`
	ConsumedUTC   chrono.Time    `json:"consumedUtc" db:"consumed_utc"` // For order kind create, renew or upgrade, an invoice is consumed immediately; for addon, it is usually consumed at a future time.
	dt.DateTimePeriod
	CarriedOverUtc chrono.Time `json:"carriedOver" db:"carried_over_utc"` // In case user has carry-over for upgrading or switching stripe, add a timestamp to original invoice.
}

// NewAddonInvoice creates a new addon invoice based on
// the parameters provided in a request.
func NewAddonInvoice(params AddOnParams) Invoice {
	return Invoice{
		ID:             ids.InvoiceID(),
		CompoundID:     params.CompoundID,
		Edition:        params.Edition,
		YearMonthDay:   params.YearMonthDay,
		AddOnSource:    params.AddOnSource,
		AppleTxID:      null.String{},
		OrderID:        params.OrderID,
		OrderKind:      enum.OrderKindAddOn,
		PaidAmount:     params.PaidAmount,
		PaymentMethod:  params.PaymentMethod,
		PriceID:        params.PriceID,
		StripeSubsID:   null.String{},
		CreatedUTC:     chrono.TimeNow(),
		ConsumedUTC:    chrono.Time{},
		DateTimePeriod: dt.DateTimePeriod{},
		CarriedOverUtc: chrono.Time{},
	}
}

func (i Invoice) SetPeriod(start time.Time) Invoice {

	if start.IsZero() {
		return i
	}

	period := dt.NewTimeRange(start).
		WithPeriod(i.YearMonthDay).
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

func (i Invoice) IsCarriedOverByUpgrade() bool {
	return i.AddOnSource == addon.SourceCarryOver && i.OrderID.Valid
}

func (i Invoice) IsCarriedOverByStripe() bool {
	return i.AddOnSource == addon.SourceCarryOver && i.StripeSubsID.Valid
}

func (i Invoice) IsCarriedOverByApple() bool {
	return i.AddOnSource == addon.SourceCarryOver && i.AppleTxID.Valid
}

func (i Invoice) IsZero() bool {
	return i.ID == ""
}

// WithOrderID set the order id field for carried-over invoice caused by upgrading.
func (i Invoice) WithOrderID(id string) Invoice {
	i.OrderID = null.StringFrom(id)
	return i
}

// WithStripeSubsID adds Stripe subscription id when a carried-over invoice
// is caused by switching to Stripe.
func (i Invoice) WithStripeSubsID(id string) Invoice {
	i.StripeSubsID = null.StringFrom(id)
	return i
}

// WithAppleTxID adds IAP original transaction id when a carried-over invoice
// is caused by switching to IAP.
func (i Invoice) WithAppleTxID(id string) Invoice {
	i.AppleTxID = null.StringFrom(id)
	return i
}

type List struct {
	Total int64 `json:"total" db:"row_count"`
	gorest.Pagination
	Data []Invoice `json:"data"`
	Err  error     `json:"-"`
}
