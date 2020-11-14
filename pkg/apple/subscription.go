package apple

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/product"
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
	product.Edition
	AutoRenewal bool        `json:"autoRenewal" db:"auto_renewal"`
	CreatedUTC  chrono.Time `json:"createdUtc" db:"created_utc"`
	UpdatedUTC  chrono.Time `json:"updatedUtc" db:"updated_utc"`
	FtcUserID   null.String `json:"ftcUserId" db:"ftc_user_id"` // This field is only manipulated upon link/unlink. We won't touch it upon insert/create.
	InUse       bool        `json:"inUse" db:"in_use"`          // If this subscription is being used by membership.
}

// NewSubscription builds a subscription for a user based on
// the receipt information available.
// Returns Subscription or error if the corresponding product is not found for this transaction.
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

// PermitLink checks whether a Subscription is allowed to link
// to an ftc  account. It is allowed only when the FtcUserID
// field is still empty, or is the same as the target id.
func (s Subscription) PermitLink(ftcID string) bool {

	if s.FtcUserID.IsZero() || s.FtcUserID.String == ftcID {
		return true
	}

	return false
}

func (s Subscription) ShouldUpdate(m reader.Membership) bool {
	if !m.IsIAP() {
		return false
	}

	if s.ExpiresDateUTC.Before(m.ExpireDate.Time) {
		return false
	}

	if s.ExpiresDateUTC.Equal(m.ExpireDate.Time) {
		return false
	}

	return true
}

// Membership build ftc's membership based on subscription
// from apple if this subscription is not linked
// to an ftc account.
func (s Subscription) NewMembership(id reader.MemberID) reader.Membership {
	m := reader.Membership{
		MemberID: id,
		Edition: product.Edition{
			Tier:  s.Tier,
			Cycle: s.Cycle,
		},
		ExpireDate:    chrono.DateFrom(s.ExpiresDateUTC.Time),
		PaymentMethod: enum.PayMethodApple,
		FtcPlanID:     null.String{},
		StripeSubsID:  null.String{},
		StripePlanID:  null.String{},
		AutoRenewal:   s.AutoRenewal,
		Status:        enum.SubsStatusNull,
		AppleSubsID:   null.StringFrom(s.OriginalTransactionID),
		B2BLicenceID:  null.String{},
	}

	return m
}

// BuildOn updates an existing IAP membership based on this
// transaction.
// Used when updating an existing subscription and its optionally
// linked membership.
func (s Subscription) BuildOn(m reader.Membership) reader.Membership {
	m.Tier = s.Tier
	m.Cycle = s.Cycle
	m.ExpireDate = chrono.DateFrom(s.ExpiresDateUTC.Time)
	m.PaymentMethod = enum.PayMethodApple
	m.FtcPlanID = null.String{}
	m.StripeSubsID = null.String{}
	m.StripePlanID = null.String{}
	m.AutoRenewal = s.AutoRenewal
	m.Status = enum.SubsStatusNull
	m.AppleSubsID = null.StringFrom(s.OriginalTransactionID)
	m.B2BLicenceID = null.String{}

	return m
}

type SubsList struct {
	Total int64 `json:"total" db:"row_count"`
	gorest.Pagination
	Data []Subscription `json:"data"`
	Err  error          `json:"-"`
}
