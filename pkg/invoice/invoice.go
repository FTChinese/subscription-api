package invoice

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/pkg/addon"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/guregu/null"
	"time"
)

const (
	daysOfYear  = 366
	daysOfMonth = 31
)

// Invoice is the result of order successfully paid.
// For one-time purchase upgrading or switching from one-time purchase to Stripe, an invoice of
type Invoice struct {
	ID         string `json:"id" db:"id"`
	CompoundID string `json:"compoundId" db:"compound_id"`
	price.Edition
	dt.YearMonthDay
	AddOnSource   addon.Source   `json:"addOnSource" db:"add_on_source"` // Only exists when OrderKind is AddOn.
	OrderID       null.String    `json:"orderId" db:"order_id"`          // Carry over for stripe does not have order id.
	OrderKind     enum.OrderKind `json:"orderKind" db:"order_kind"`      // Always non-null
	PaidAmount    float64        `json:"paidAmount" db:"paid_amount"`
	PaymentMethod enum.PayMethod `json:"payMethod" db:"payment_method"`
	PriceID       null.String    `json:"priceId" db:"price_id"`
	CreatedUTC    chrono.Time    `json:"createdUtc" db:"created_utc"`
	ConsumedUTC   chrono.Time    `json:"consumedUtc" db:"consumed_utc"`
	dt.DateTimePeriod
	CarriedOverUtc chrono.Time `json:"carriedOver" db:"carried_over_utc"` // In case user has carry-over for upgrading or switching stripe, add a timestamp to original invoice.
}

func NewFromCarryOver(m reader.Membership, source addon.Source) Invoice {
	return Invoice{
		ID:         db.InvoiceID(),
		CompoundID: m.CompoundID,
		Edition:    m.Edition,
		YearMonthDay: dt.YearMonthDay{
			Days: m.RemainingDays(),
		},
		AddOnSource:    source,
		OrderID:        null.String{},
		OrderKind:      enum.OrderKindAddOn, // All carry-over invoice are add-ons
		PaidAmount:     0,
		PaymentMethod:  m.PaymentMethod,
		PriceID:        m.FtcPlanID,
		CreatedUTC:     chrono.TimeNow(),
		ConsumedUTC:    chrono.Time{}, // Will be consumed in the future.
		DateTimePeriod: dt.DateTimePeriod{},
		CarriedOverUtc: chrono.Time{},
	}
}

func NewFromUpgradeCarryOver(m reader.Membership) Invoice {
	return NewFromCarryOver(m, addon.SourceUpgradeCarryOver)
}

func NewFromOneTimeToSubCarryOver(m reader.Membership) Invoice {
	return NewFromCarryOver(m, addon.SourceOneTimeToSubCarryOver)
}

func (i Invoice) SetPeriod(start time.Time) Invoice {
	period := dt.NewTimeRange(start).
		WithDate(i.YearMonthDay).
		ToDateTimePeriod()

	i.DateTimePeriod = period

	return i
}

func (i Invoice) IsZero() bool {
	return i.ID == ""
}

func (i Invoice) IsCarryOverAddOn() bool {
	return i.OrderKind == enum.OrderKindAddOn && (i.AddOnSource == addon.SourceUpgradeCarryOver || i.AddOnSource == addon.SourceOneTimeToSubCarryOver)
}

func (i Invoice) IsCompensationAddOn() bool {
	return i.OrderKind == enum.OrderKindAddOn && i.AddOnSource == addon.SourceCompensation
}

func (i Invoice) IsPurchasedAddOn() bool {
	return i.OrderKind == enum.OrderKindAddOn && i.AddOnSource == ""
}

// GetDays calculates roughly the how many days this add-on has.
// It is not precise and used only as an indicator that user has add-on.
func (i Invoice) GetDays() int64 {
	return i.Years*daysOfYear + i.Months*daysOfMonth + i.Days
}

// ToReservedDays calculates how many days this add-on could be converted to reserved part of membership.
func (i Invoice) ToReservedDays() addon.ReservedDays {
	switch i.Tier {
	case enum.TierStandard:
		return addon.ReservedDays{
			Standard: i.GetDays(),
			Premium:  0,
		}
	case enum.TierPremium:
		return addon.ReservedDays{
			Standard: 0,
			Premium:  i.GetDays(),
		}

	// Returns zero if current instance is zero.
	default:
		return addon.ReservedDays{}
	}
}

func (i Invoice) WithOrderID(id string) Invoice {
	i.OrderID = null.StringFrom(id)
	return i
}
