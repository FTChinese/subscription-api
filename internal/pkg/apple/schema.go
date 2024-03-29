package apple

import (
	"github.com/guregu/null"
)

// BaseSchema contains the shared fields of all schema.
type BaseSchema struct {
	Environment           Environment `json:"environment" db:"environment"`
	OriginalTransactionID string      `json:"originalTransactionId" db:"original_transaction_id"`
}

// ReceiptFileName builds the file name when persisting latest receipt to disk.
func (s BaseSchema) ReceiptFileName() string {
	return s.OriginalTransactionID + "_" + s.Environment.String() + ".txt"
}

// ReceiptKeyName creates key used in redis: `iap:receipt:1000000922681985-Sandbox`
func (s BaseSchema) ReceiptKeyName() string {
	return "iap:receipt:" + s.OriginalTransactionID + "-" + s.Environment.String()
}

// VerifiedReceiptSchema is the SQL version of ClientReceipt.
type VerifiedReceiptSchema struct {
	BaseSchema
	TransactionID              string      `db:"transaction_id"`
	AppItemID                  int64       `db:"app_item_id"`
	ApplicationVersion         string      `db:"application_version"`
	BundleID                   string      `db:"bundle_id"`
	DownloadID                 int64       `db:"download_id"`
	ExpirationDateMs           null.Int    `db:"expiration_date_ms"`
	OriginalPurchaseDateMs     int64       `db:"original_purchase_date_ms"`
	OriginalApplicationVersion string      `db:"original_application_version"`
	PreorderDateMs             null.Int    `db:"preorder_date_ms"`
	ReceiptCreationDateMs      int64       `db:"receipt_creation_date_ms"`
	ReceiptType                ReceiptType `db:"receipt_type"`
	RequestDateMs              int64       `db:"request_date_ms"`
	Status                     int64       `db:"status"`
	VersionExternalIdentifier  int64       `db:"version_external_identifier"`
}

type BaseTransactionSchema struct {
	BaseSchema
	TransactionID string `db:"transaction_id"` // UNIQUE INDEX

	ExpiresDateMs          int64  `db:"expires_date_ms"`
	IsInIntroOfferPeriod   bool   `db:"is_in_intro_offer_period"`
	IsTrialPeriod          bool   `db:"is_trial_period"`
	OriginalPurchaseDateMs int64  `db:"original_purchase_date_ms"`
	ProductID              string `db:"product_id"`
	PurchaseDateMs         int64  `db:"purchase_date_ms"`
	Quantity               int64  `db:"quantity"`
	WebOrderLineItemID     string `db:"web_order_line_item_id"`
}

// TransactionSchema represents the db schema to select/insert a row of receipt.
// Use OriginalTransactionID, ProductID, ExpiresDateMs and PurchaseDateMs to identify auto-renewable subscriptions.
type TransactionSchema struct {
	BaseTransactionSchema

	CancellationDateMs  null.Int    `db:"cancellation_date_ms"`
	CancellationReason  null.String `db:"cancellation_reason"`
	IsUpgraded          bool        `db:"is_upgraded"`
	PromotionalOfferID  null.String `db:"promotional_offer_id"`
	SubscriptionGroupID null.String `db:"sub_group_id"`
}

type PendingRenewalSchema struct {
	BaseSchema

	AutoRenewProductID       null.String `db:"auto_renew_product_id"`
	AutoRenewStatus          string      `db:"auto_renew_status"`
	ExpirationIntent         null.String `db:"expiration_intent"`
	GracePeriodExpiresDateMs null.Int    `db:"grace_period_expires_date_ms"`
	IsInBillingRetryPeriod   null.Bool   `db:"is_in_billing_retry_period"`
	PriceConsentStatus       null.String `db:"price_consent_status"`
	ProductID                string      `db:"product_id"`
}
