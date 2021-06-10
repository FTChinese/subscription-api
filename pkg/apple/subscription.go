package apple

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/guregu/null"
	"time"
)

type SubsFilter struct {
	ExpiresIn   int64  // expires_date_utc
	AutoRenewal bool   // Allow null
	OrderBy     string // Order by column: updated_utc, expires_date_utc
	Ascending   bool
}

// Subscription contains a user's subscription data.
// It it built from app store's verification response.
// The original transaction id is used to uniquely identify a user.
type Subscription struct {
	BaseSchema
	LastTransactionID string      `json:"lastTransactionId" db:"last_transaction_id"`
	ProductID         string      `json:"productId" db:"product_id"`
	PurchaseDateUTC   chrono.Time `json:"purchaseDateUtc" db:"purchase_date_utc"`
	ExpiresDateUTC    chrono.Time `json:"expiresDateUtc" db:"expires_date_utc"`
	price.Edition
	AutoRenewal bool        `json:"autoRenewal" db:"auto_renewal"`
	CreatedUTC  chrono.Time `json:"createdUtc" db:"created_utc"`
	UpdatedUTC  chrono.Time `json:"updatedUtc" db:"updated_utc"`
	FtcUserID   null.String `json:"ftcUserId" db:"ftc_user_id"` // This field is only manipulated upon link/unlink. We won't touch it upon insert/create.
	InUse       bool        `json:"inUse" db:"in_use"`          // If this subscription is being used by membership.
}

// NewSubscription builds a subscription for a user based on
// the receipt information available.
// Returns Subscription or error if the corresponding price is not found for this transaction.
// When we build a new Subscription from apple verification response,
// we do no know user's ftc id,  so leave it empty.
// And do not touch the ftc_user_id field when you inserting/updating a Subscription.
func NewSubscription(u UnifiedReceipt) (Subscription, error) {
	pendingRenewal := u.findPendingRenewal()

	autoRenew := pendingRenewal.IsAutoRenew()
	if u.latestTransaction.IsCancelled() {
		autoRenew = false
	}

	prod, err := appleProducts.findByID(u.latestTransaction.ProductID)

	if err != nil {
		return Subscription{}, err
	}

	return Subscription{
		BaseSchema: BaseSchema{
			Environment:           u.Environment,
			OriginalTransactionID: u.latestTransaction.OriginalTransactionID,
		},
		LastTransactionID: u.latestTransaction.TransactionID,
		ProductID:         u.latestTransaction.ProductID,
		PurchaseDateUTC: chrono.TimeFrom(
			time.Unix(u.latestTransaction.PurchaseDateUnix(), 0),
		),
		ExpiresDateUTC: chrono.TimeFrom(
			time.Unix(u.latestTransaction.ExpiresUnix(), 0),
		),
		Edition:     prod.Edition,
		AutoRenewal: autoRenew,
		CreatedUTC:  chrono.TimeNow(),
		UpdatedUTC:  chrono.TimeNow(),
	}, nil
}

func (s Subscription) IsExpired() bool {
	if s.AutoRenewal {
		return false
	}

	return s.ExpiresDateUTC.Before(time.Now())
}

// PermitLink checks whether a Subscription is allowed to link
// to an ftc  account. It is allowed only when the FtcUserID
// field is still empty, or is the same as the target id.
func (s Subscription) PermitLink(ftcID string) bool {

	if s.FtcUserID.IsZero() || s.FtcUserID.String == ftcID {
		return true
	}

	return false
}

// ShouldUpdate checks if the subscription modified.
// If not, we won't touch db for membership.
func (s Subscription) ShouldUpdate(m reader.Membership) bool {
	if !m.IsIAP() {
		return false
	}

	if !s.ExpiresDateUTC.After(m.ExpireDate.Time) {
		return false
	}

	return true
}

type SubsList struct {
	Total int64 `json:"total" db:"row_count"`
	gorest.Pagination
	Data []Subscription `json:"data"`
	Err  error          `json:"-"`
}
