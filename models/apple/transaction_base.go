package apple

type BaseTransaction struct {
	// An indicator of whether an auto-renewable subscription is in the introductory price period.
	// true The customer’s subscription is in an introductory price period
	// false The subscription is not in an introductory price period.
	IsInIntroOfferPeriod string `json:"is_in_intro_offer_period"`
	// An indicator of whether a subscription is in the free trial period.
	// true The subscription is in the free trial period.
	// false The subscription is not in the free trial period.
	IsTrialPeriod string `json:"is_trial_period"`

	// The time of the original app purchase
	OriginalPurchaseDate    string `json:"original_purchase_date"`
	OriginalPurchaseDateMs  string `json:"original_purchase_date_ms"`
	OriginalPurchaseDatePST string `json:"original_purchase_date_pst"`

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
	OriginalTransactionID string `json:"original_transaction_id"`
	TransactionID         string `json:"transaction_id" db:"transaction_id"`

	ProductID string `json:"product_id"`

	// The time the App Store charged the user’s account for a subscription purchase or renewal after a lapse
	PurchaseDate    string `json:"purchase_date"`
	PurchaseDateMs  string `json:"purchase_date_ms"`
	PurchaseDatePST string `json:"purchase_date_pst"`

	// The number of consumable products purchased.
	// The value is usually “1” unless modified with a mutable payment. The maximum value is "10".
	Quantity string `json:"quantity"`

	// A unique identifier for purchase events across devices, including subscription-renewal events. This value is the primary key for identifying subscription purchases.
	WebOrderLineItemID string `json:"web_order_line_item_id"`
}

// Schema turns to a DB schema structure.
func (t BaseTransaction) schema(e Environment, expires string) BaseTransactionSchema {
	return BaseTransactionSchema{
		BaseSchema: BaseSchema{
			Environment:           e,
			OriginalTransactionID: t.OriginalTransactionID,
		},
		TransactionID:          t.TransactionID,
		ExpiresDateMs:          MustParseInt64(expires),
		IsInIntroOfferPeriod:   MustParseBoolean(t.IsInIntroOfferPeriod),
		IsTrialPeriod:          MustParseBoolean(t.IsTrialPeriod),
		OriginalPurchaseDateMs: MustParseInt64(t.OriginalPurchaseDateMs),
		ProductID:              t.ProductID,
		PurchaseDateMs:         MustParseInt64(t.PurchaseDateMs),
		Quantity:               MustParseInt64(t.Quantity),
		WebOrderLineItemID:     t.WebOrderLineItemID,
	}
}
