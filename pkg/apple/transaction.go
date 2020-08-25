package apple

import (
	"github.com/guregu/null"
)

// Transaction is the action performed during subscription.
// Apple calls this as receipt. But it is not correct.
// This is just a snapshot, or history of transactions.
type Transaction struct {
	// The time Apple customer support canceled a transaction
	// This field is only present for refunded transactions.
	CancellationDate string `json:"cancellation_date"`
	// The time Apple customer support canceled a transaction, or the time an auto-renewable subscription plan was upgraded
	CancellationDateMs  string `json:"cancellation_date_ms"`
	CancellationDatePST string `json:"cancellation_date_pst"`
	// The reason for a refunded transaction.
	// When a customer cancels a transaction, the App Store gives them a refund and provides a value for this key.
	// “1” indicates that the customer canceled their transaction due to an actual or perceived issue within your app.
	// “0” indicates that the transaction was canceled for another reason; for example, if the customer made the purchase accidentally.
	CancellationReason string `json:"cancellation_reason"`

	// The time a subscription expires or when it will renew
	ExpiresDate    string `json:"expires_date"`
	ExpiresDateMs  string `json:"expires_date_ms"`
	ExpiresDatePST string `json:"expires_date_pst"`

	BaseTransaction

	// An indicator that a subscription has been canceled due to an upgrade. This field is only present for upgrade transactions.
	IsUpgraded          string      `json:"is_upgraded"`
	PromotionalOfferID  null.String `json:"promotional_offer_id"`
	SubscriptionGroupID null.String `json:"subscription_group_identifier"` // This does not exist for `in_app` field.
}

func (t Transaction) IsCancelled() bool {
	return t.CancellationDateMs != ""
}

// CancellationUnix turns cancellation milliseconds in string
// to Unix seconds.
// Returns 0 if it is not cancelled.
func (t Transaction) CancellationUnix() int64 {
	return MustParseInt64(t.CancellationDateMs) / 1000
}

// IsValidProduct checks if the product exists.
func (t Transaction) IsValidProduct() bool {
	return appleProducts.exists(t.ProductID)
}

// ExpiresUnix turns the milliseconds in string to a unix timestamp in seconds.
func (t Transaction) ExpiresUnix() int64 {
	return MustParseInt64(t.ExpiresDateMs) / 1000
}

// PurchaseDateUnix turns the milliseconds string to unix seconds.
func (t Transaction) PurchaseDateUnix() int64 {
	return MustParseInt64(t.PurchaseDateMs) / 1000
}

// Schema turns a Transaction to a database schema structure.
func (t Transaction) Schema(e Environment) TransactionSchema {

	cd := MustParseInt64(t.CancellationDateMs)

	return TransactionSchema{

		BaseTransactionSchema: t.schema(
			e,
			t.ExpiresDateMs,
		),
		IsUpgraded:          MustParseBoolean(t.IsUpgraded),
		SubscriptionGroupID: t.SubscriptionGroupID,
		CancellationDateMs:  null.NewInt(cd, cd != 0),
		CancellationReason:  null.StringFrom(t.CancellationReason),
		PromotionalOfferID:  t.PromotionalOfferID,
	}
}
