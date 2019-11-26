package apple

import "github.com/guregu/null"

// ReceiptInfo is the server-side receipt.
type ReceiptInfo struct {
	CancellationDate       null.String `json:"cancellation_date"`    // The time Apple customer support canceled a transaction
	CancellationDateMs     null.String `json:"cancellation_date_ms"` // The time Apple customer support canceled a transaction, or the time an auto-renewable subscription plan was upgraded, in UNIX epoch time format, in milliseconds.
	CancellationDatePST    null.String `json:"cancellation_date_pst"`
	CancellationReason     null.String `json:"cancellation_reason"` // The reason for a refunded transaction. “1” indicates that the customer canceled their transaction due to an actual or perceived issue within your app. “0” indicates that the transaction was canceled for another reason; for example, if the customer made the purchase accidentally.
	ExpiresDate            string      `json:"expires_date"`        // The time a subscription expires or when it will renew, in a date-time format similar to the ISO 8601.
	ExpiresDateMs          string      `json:"expires_date_ms"`     // The time a subscription expires or when it will renew, in UNIX epoch time format, in milliseconds.
	ExpiresDatePST         string      `json:"expires_date_pst"`
	IsInIntroOfferPeriod   Boolean     `json:"is_in_intro_offer_period"` // true | false
	IsTrialPeriod          Boolean     `json:"is_trial_period"`          // true | false
	IsUpgraded             Boolean     `json:"is_upgraded"`
	OriginalPurchaseDate   string      `json:"original_purchase_date"`
	OriginalPurchaseDateMs string      `json:"original_purchase_date_ms"`
	OriginalPurchasePST    string      `json:"original_purchase_pst"`
	OriginalTransactionID  string      `json:"original_transaction_id"`
	ProductID              string      `json:"product_id"`
	PromotionalOfferID     null.String `json:"promotional_offer_id"`
	PurchaseDate           string      `json:"purchase_date"`
	PurchaseDateMs         string      `json:"purchase_date_ms"`
	PurchaseDatePST        string      `json:"purchase_date_pst"`
	Quantity               string      `json:"quantity"`
	SubscriptionGroupID    null.String `json:"subscription_group_identifier"` // This does not exist for `in_app` field.
	TransactionID          string      `json:"transaction_id"`
	WebOrderLineItemID     string      `json:"web_order_line_item_id"` // A unique identifier for purchase events across devices, including subscription-renewal events. This value is the primary key for identifying subscription purchases.
}

// PendingRenewal contains auto-renewable subscription renewals that are open or failed in the past.
type PendingRenewal struct {
	AutoRenewalProductID      string `json:"auto_renewal_product_id"`
	AutoRenewalStatus         string `json:"auto_renewal_status"` // 1: The subscription will renew at the end of the current subscription period. 0: The customer has turned off automatic renewal for the subscription.
	ExpirationIntent          string `json:"expiration_intent"`   // 1 The customer voluntarily canceled their subscription. 2 Billing error; for example, the customer's payment information was no longer valid. 3 The customer did not agree to a recent price increase. 4 The product was not available for purchase at the time of renewal. 5 Unknown error.
	GracePeriodExpiresDate    string `json:"grace_period_expires_date"`
	GracePeriodExpiresDateMs  string `json:"grace_period_expires_date_ms"`
	GracePeriodExpiresDatePST string `json:"grace_period_expires_date_pst"`
	IsInBillingRetryPeriod    string `json:"is_in_billing_retry_period"`
	OriginalTransactionID     string `json:"original_transaction_id"`
	PriceConsentStatus        string `json:"price_consent_status"`
	ProductID                 string `json:"product_id"`
}

// DecodedReceipt is the decoded version of the encoded receipt data sent with the request to the App Store.
type DecodedReceipt struct {
	AdamID             int64       `json:"adam_id"`
	AppItemID          int64       `json:"app_item_id"` // uniquely identify the app purchased. only in production. O for sandbox.
	ApplicationVersion string      `json:"application_version"`
	BundleID           string      `json:"bundle_id"`
	DownloadID         string      `json:"download_id"`     // A unique identifier for the app download transaction.
	ExpirationDate     null.String `json:"expiration_date"` // The time the receipt expires for apps purchased through the Volume Purchase Program, short for VPP.
	ExpirationDateMs   null.String `json:"expiration_date_ms"`
	ExpirationDatePST  null.String `json:"expiration_date_pst"`
	// An array that contains the in-app purchase receipt fields for all in-app purchase transactions.
	// It is not in chronological order. When parsing the array, iterate over all items to ensure all items are fulfilled
	// Use this array to:
	// Check for an empty array in a valid receipt, indicating that the App Store has made no in-app purchase charges.
	// Determine which products the user purchased. Purchases for non-consumable products, auto-renewable subscriptions, and non-renewing subscriptions remain in the receipt indefinitely.
	InApp                      []ReceiptInfo `json:"in_app"`
	OriginalApplicationVersion string        `json:"original_application_version"` // The version of the app that the user originally purchased
	OriginalPurchaseDate       string        `json:"original_purchase_date"`
	OriginalPurchaseDateMs     string        `json:"original_purchase_date_ms"`
	PreorderDate               null.String   `json:"preorder_date"` // The time the user ordered the app available for pre-order
	PreorderDateMs             null.String   `json:"preorder_date_ms"`
	PreorderDatePST            null.String   `json:"preorder_date_pst"`
	ReceiptCreationDate        string        `json:"receipt_creation_date"` // The time the App Store generated the receipt
	ReceiptCreationDateMs      string        `json:"receipt_creation_date_ms"`
	ReceiptCreationDatePST     string        `json:"receipt_creation_date_pst"`
	ReceiptType                string        `json:"receipt_type"` // Production | ProductionVPP | ProductionSandbox | ProductionVPPSandbox
	RequestDate                string        `json:"request_date"` // The time the request to the verifyReceipt endpoint was processed and the response was generated
	RequestDateMs              string        `json:"request_date_ms"`
	RequestDatePST             string        `json:"request_date_pst"`
	VersionExternalIdentifier  int64         `json:"version_external_identifier"` // An arbitrary number that identifies a revision of your app. In the sandbox, this key's value is “0”.
}
