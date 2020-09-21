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

// Create key used in redis: `iap:receipt:1000000922681985-Sandbox`
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

// ReceiptToken is a row to save the receipt data from a request
// so that later we can re-validate this receipt.
type ReceiptToken struct {
	BaseSchema
	LatestReceipt string `db:"latest_receipt"`
}

// WebHookSchema saves the value of WebHook root fields and the values of its LatestTransaction fields.
type WebHookSchema struct {
	BaseTransactionSchema
	AppItemID int64 `db:"app_item_id"`
	ItemID    int64 `db:"item_id"`

	// Root elements
	AutoRenewAdamID             int64            `db:"auto_renew_adam_id"`
	AutoRenewProductID          string           `db:"auto_renew_product_id"`
	AutoRenewStatus             null.Bool        `db:"auto_renew_status"`
	AutoRenewStatusChangeDateMs int64            `db:"auto_renew_status_change_date_ms"`
	ExpirationIntent            null.String      `db:"expiration_intent"`
	NotificationType            NotificationType `db:"notification_type"`
	Password                    string           `db:"password"`
	Status                      int64            `db:"status"`
}

// Save the receipt as a token for status polling.
const StmtSaveReceiptToken = `
INSERT INTO premium.apple_receipt_token
SET environment = :environment,
	original_transaction_id = :original_transaction_id,
	latest_receipt = :latest_receipt,
	updated_utc = UTC_TIMESTAMP(),
	created_utc = UTC_TIMESTAMP()
ON DUPLICATE KEY UPDATE
	latest_receipt = :latest_receipt,
	updated_utc = UTC_TIMESTAMP()`
