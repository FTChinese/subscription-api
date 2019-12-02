package apple

import (
	"github.com/guregu/null"
	"gitlab.com/ftchinese/subscription-api/models/plan"
)

// ReceiptInfo is the server-side receipt.
type ReceiptInfo struct {
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

	IsInIntroOfferPeriod string `json:"is_in_intro_offer_period"` // true | false
	IsTrialPeriod        string `json:"is_trial_period"`          // true | false
	IsUpgraded           string `json:"is_upgraded"`

	OriginalPurchaseDate   string `json:"original_purchase_date"`
	OriginalPurchaseDateMs string `json:"original_purchase_date_ms"`
	OriginalPurchasePST    string `json:"original_purchase_pst"`

	// The transaction identifier of the original purchase
	// This value is identical to the transaction_id except when the user restores a purchase or renews a subscription.
	// For auto-renewable subscription transactions, this value also appears in the pending_renewal_info
	// use this value to:
	// * Match a transaction found in the receipt to a server-to-server notification event
	// * Manage auto-renewable subscriptions. Store this value, product_id, expires_date_ms, and purchase_date_ms for each transaction for each customer
	// * Identify a subscription with the product_id in the pending_renewal_info section. Treat this purchase as a new subscription when you see a new original_transaction_id value for a product_id
	// * Differentiate a purchase transaction from a restore or a renewal transaction.
	//   * In a purchase transaction, the transaction_id always matches the original_transaction_id.
	//   * For subscriptions, it indicates the first subscription purchase
	//   * For a restore or renewal, the transaction_id does not match the original_transaction_id.
	// * Identify one or more renewals for the same subscription
	OriginalTransactionID string      `json:"original_transaction_id"`
	ProductID             string      `json:"product_id" db:"product_id"`
	PromotionalOfferID    null.String `json:"promotional_offer_id"`
	PurchaseDate          string      `json:"purchase_date"`
	PurchaseDateMs        string      `json:"purchase_date_ms" db:"purchase_date_ms"`
	PurchaseDatePST       string      `json:"purchase_date_pst"`
	Quantity              string      `json:"quantity" db:"quantity"`
	SubscriptionGroupID   null.String `json:"subscription_group_identifier"` // This does not exist for `in_app` field.
	TransactionID         string      `json:"transaction_id" db:"transaction_id"`
	// A unique identifier for purchase events across devices, including subscription-renewal events. This value is the primary key for identifying subscription purchases.
	WebOrderLineItemID string `json:"web_order_line_item_id"`
}

func (r ReceiptInfo) IsCancelled() bool {
	return r.CancellationDateMs != ""
}

func (r ReceiptInfo) IsValidProduct() bool {
	return plan.AppleProductExists(r.ProductID)
}

// ExpiresUnix turns the milliseconds in string to a unix timestamp in seconds.
func (r ReceiptInfo) ExpiresUnix() int64 {
	return MustParseInt64(r.ExpiresDateMs) / 1000
}

func (r ReceiptInfo) PurchaseDateUnix() int64 {
	return MustParseInt64(r.PurchaseDateMs) / 1000
}

func (r ReceiptInfo) FindPlan() plan.Plan {
	p, _ := plan.FindPlanForApple(r.ProductID)

	return p
}

func (r ReceiptInfo) Schema(e Environment) ReceiptSchema {

	cd := MustParseInt64(r.CancellationDateMs)

	return ReceiptSchema{
		Environment:            e,
		Quantity:               MustParseInt64(r.Quantity),
		ProductID:              r.ProductID,
		TransactionID:          r.TransactionID,
		OriginalTransactionID:  r.OriginalTransactionID,
		PurchaseDateMs:         MustParseInt64(r.PurchaseDateMs),
		OriginalPurchaseDateMs: MustParseInt64(r.OriginalPurchaseDateMs),
		ExpiresDateMs:          MustParseInt64(r.ExpiresDateMs),
		WebOrderLineItemID:     r.WebOrderLineItemID,
		IsInIntroOfferPeriod:   MustParseBoolean(r.IsInIntroOfferPeriod),
		IsTrialPeriod:          MustParseBoolean(r.IsTrialPeriod),
		IsUpgraded:             MustParseBoolean(r.IsUpgraded),
		SubscriptionGroupID:    r.SubscriptionGroupID,
		CancellationDateMs:     null.NewInt(cd, cd != 0),
		CancellationReason:     null.StringFrom(r.CancellationReason),
		PromotionalOfferID:     r.PromotionalOfferID,
	}
}
