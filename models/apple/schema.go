package apple

import (
	"github.com/guregu/null"
)

// VerificationSessionSchema contains the meta data of verification result.
type VerificationSessionSchema struct {
	Status                     int64       `db:"status"`
	Environment                Environment `db:"environment"`
	OriginalTransactionID      string      `db:"original_transaction_id"`
	TransactionID              string      `db:"transaction_id"`
	ReceiptType                ReceiptType `db:"receipt_type"`
	AppItemID                  int64       `db:"app_item_id"`
	BundleID                   string      `db:"bundle_id"`
	ApplicationVersion         string      `db:"application_version"`
	DownloadID                 int64       `db:"download_id"`
	VersionExternalIdentifier  int64       `db:"version_external_identifier"`
	ReceiptCreationDateMs      int64       `db:"receipt_creation_date_ms"`
	RequestDateMs              int64       `db:"request_date_ms"`
	OriginalPurchaseDateMs     int64       `db:"original_purchase_date_ms"`
	ExpirationDateMs           null.Int    `db:"expiration_date_ms"`
	OriginalApplicationVersion string      `db:"original_application_version"`
	PreorderDateMs             null.Int    `db:"preorder_date_ms"`
}

type VerificationFailed struct {
	Environment Environment `db:"environment"`
	Status      int64       `db:"status"`
	Message     null.String `db:"message"`
	ReceiptData string      `db:"receipt_data"`
}

// ReceiptRow represents the db schema to select/insert a row of receipt.
// Use OriginalTransactionID, ProductID, ExpiresDateMs and PurchaseDateMs to identify auto-renewable subscriptions.
type ReceiptSchema struct {
	Environment Environment `db:"environment"`
	Quantity    int64       `db:"quantity"`

	ProductID              string `db:"product_id"`
	TransactionID          string `db:"transaction_id"`
	OriginalTransactionID  string `db:"original_transaction_id"`
	PurchaseDateMs         int64  `db:"purchase_date_ms"`
	OriginalPurchaseDateMs int64  `db:"original_purchase_date_ms"`
	ExpiresDateMs          int64  `db:"expires_date_ms"`

	WebOrderLineItemID   string      `db:"web_order_line_item_id"`
	IsInIntroOfferPeriod bool        `db:"is_in_intro_offer_period"`
	IsTrialPeriod        bool        `db:"is_trial_period"`
	IsUpgraded           bool        `db:"is_upgraded"`
	SubscriptionGroupID  null.String `db:"sub_group_id"`
	CancellationDateMs   null.Int    `db:"cancellation_date_ms"`
	CancellationReason   null.String `db:"cancellation_reason"`
	PromotionalOfferID   null.String `db:"promotional_offer_id"`
}

type PendingRenewalSchema struct {
	Environment              Environment `db:"environment"`
	OriginalTransactionID    string      `db:"original_transaction_id"`
	ProductID                string      `db:"product_id"`
	AutoRenewStatus          string      `db:"auto_renew_status"`
	ExpirationIntent         null.String `db:"expiration_intent"`
	AutoRenewProductID       null.String `db:"auto_renew_product_id"`
	IsInBillingRetryPeriod   null.Bool   `db:"is_in_billing_retry_period"`
	GracePeriodExpiresDateMs null.Int    `db:"grace_period_expires_date_ms"`
	PriceConsentStatus       null.String `db:"price_consent_status"`
}

// ReceiptToken is a row to save the receipt data from a request
// so that later we can re-validate this receipt.
type ReceiptToken struct {
	Environment           Environment `db:"environment"`
	OriginalTransactionID string      `db:"original_transaction_id"`
	LatestReceipt         string      `db:"latest_receipt"`
}
