package apple

import (
	"github.com/guregu/null"
	"gitlab.com/ftchinese/subscription-api/models/paywall"
	"strconv"
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
	return paywall.AppleProductExists(r.ProductID)
}

// ExpiresUnix turns the milliseconds in string to a unix timestamp in seconds.
func (r ReceiptInfo) ExpiresUnix() int64 {
	return MustParseInt64(r.ExpiresDateMs) / 1000
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

// PendingRenewal contains auto-renewable subscription renewals that are open or failed in the past.
type PendingRenewal struct {
	// only present if the user downgrades or crossgrades to a subscription of a different duration for the subsequent subscription period
	AutoRenewProductID null.String `json:"auto_renew_product_id"`
	// The renewal status for the auto-renewable subscription.
	// 1: The subscription will renew at the end of the current subscription period.
	// 0: The customer has turned off automatic renewal for the subscription.
	// The value for this field should not be interpreted as the customer’s subscription status
	AutoRenewStatus string `json:"auto_renew_status"`
	// The reason a subscription expired.
	// This field is only present for a receipt that contains an expired auto-renewable subscription.
	// 1 The customer voluntarily canceled their subscription.
	// 2 Billing error; for example, the customer's payment information was no longer valid.
	// 3 The customer did not agree to a recent price increase.
	// 4 The product was not available for purchase at the time of renewal.
	// 5 Unknown error.
	ExpirationIntent null.String `json:"expiration_intent"`
	// The time at which the grace period for subscription renewals expires
	GracePeriodExpiresDate string `json:"grace_period_expires_date"`
	// This key is only present for apps that have Billing Grace Period enabled and when the user experiences a billing error at the time of renewal.
	GracePeriodExpiresDateMs  null.String `json:"grace_period_expires_date_ms"`
	GracePeriodExpiresDatePST string      `json:"grace_period_expires_date_pst"`
	// A flag that indicates Apple is attempting to renew an expired subscription automatically.
	// This field is only present if an auto-renewable subscription is in the billing retry state
	// 1 - The App Store is attempting to renew the subscription.
	// 0 - The App Store has stopped attempting to renew the subscription.
	IsInBillingRetryPeriod null.String `json:"is_in_billing_retry_period"`
	// The transaction identifier of the original purchase.
	OriginalTransactionID string `json:"original_transaction_id"`
	// This field is only present if the customer was notified of the price increase.
	// The default value is "0" and changes to "1" if the customer consents.
	PriceConsentStatus null.String `json:"price_consent_status"`
	ProductID          string      `json:"product_id"`
}

func (p PendingRenewal) IsAutoRenew() null.Bool {
	ok, err := strconv.ParseBool(p.AutoRenewStatus)
	if err != nil {
		return null.Bool{}
	}

	return null.BoolFrom(ok)
}

func (p PendingRenewal) Schema(e Environment) PendingRenewalSchema {
	return PendingRenewalSchema{
		Environment:           e,
		OriginalTransactionID: p.OriginalTransactionID,
		ProductID:             p.ProductID,
		AutoRenewStatus:       p.AutoRenewStatus,
		ExpirationIntent:      p.ExpirationIntent,
		AutoRenewProductID:    p.AutoRenewProductID,
		IsInBillingRetryPeriod: null.NewBool(
			MustParseBoolean(p.IsInBillingRetryPeriod.String),
			p.IsInBillingRetryPeriod.Valid),
		GracePeriodExpiresDateMs: null.NewInt(
			MustParseInt64(p.GracePeriodExpiresDateMs.String),
			p.GracePeriodExpiresDateMs.Valid),
		PriceConsentStatus: p.PriceConsentStatus,
	}
}

// DecodedReceipt is the decoded version of the encoded receipt data sent with the request to the App Store.
type DecodedReceipt struct {
	AdamID             int64       `json:"adam_id"`
	AppItemID          int64       `json:"app_item_id"` // uniquely identify the app purchased. only in production. O for sandbox.
	ApplicationVersion string      `json:"application_version"`
	BundleID           string      `json:"bundle_id"`
	DownloadID         int64       `json:"download_id"`     // A unique identifier for the app download transaction.
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
