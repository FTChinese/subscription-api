package apple

import (
	"github.com/guregu/null"
)

type LatestTransaction struct {
	BaseTransaction

	// An identifier that App Store Connect generates and the App Store uses to uniquely identify the app purchased.
	AppItemID string `json:"app_item_id"`
	// An identifier that App Store Connect generates and the App Store uses to uniquely identify the in-app product purchased.
	ItemID string `json:"item_id"`

	// The time a subscription expires or when it will renew, in UNIX epoch time format, in milliseconds.
	ExpiresDate             string `json:"expires_date"`
	ExpiresDateFormatted    string `json:"expires_date_formatted"`
	ExpiresDateFormattedPST string `json:"expires_date_formatted_pst"`
}

// WebHook contains the JSON data sent in the server notification from the App Store.
// It seems the top level data is a mix of ClientReceipt
// and PendingRenewal.
type WebHook struct {
	// An identifier that App Store Connect generates and the App Store uses to uniquely identify the auto-renewable subscription that the user's subscription renews
	AutoRenewAdamID int64 `json:"auto_renew_adam_id"`
	// The product identifier of the auto-renewable subscription that the user's subscription renews.
	// Also present in PendingRenewal
	AutoRenewProductID string `json:"auto_renew_product_id"`
	// The current renewal status for an auto-renewable subscription product.
	// Note that these values are different from those of the auto_renew_status in the receipt.
	// true | false
	// Also present in PendingRenewal
	AutoRenewStatus string `json:"auto_renew_status"`

	// The time at which the renewal status for an auto-renewable subscription was turned on or off,
	AutoRenewStatusChangeDate    string `json:"auto_renew_status_change_date"`
	AutoRenewStatusChangeDateMs  string `json:"auto_renew_status_change_date_ms"`
	AutoRenewStatusChangeDatePST string `json:"auto_renew_status_change_date_pst"`

	// The environment for which the receipt was generated.
	//Possible values: Sandbox, PROD
	Environment Environment `json:"environment"`

	// The reason a subscription expired. This field is only present for an expired auto-renewable subscription.
	// 1 The customer voluntarily canceled their subscription.
	// 2 Billing error; for example, the customer's payment information was no longer valid.
	// 3 The customer did not agree to a recent price increase.
	// 4 The product was not available for purchase at the time of renewal.
	// 5 Unknown error.
	// Also found in PendingRenewal.
	ExpirationIntent null.String `json:"expiration_intent"`

	// The latest Base64-encoded transaction receipt
	LatestExpiredReceipt     string        `json:"latest_expired_receipt"`
	LatestExpiredReceiptInfo []Transaction `json:"latest_expired_receipt_info"`
	// It seems this one is the same as UnifiedReceipt.LatestReceipt.
	LatestToken string `json:"latest_receipt"`
	// Saved to db.
	LatestTransaction LatestTransaction `json:"latest_receipt_info"`
	NotificationType  NotificationType  `json:"notification_type"`
	Password          string            `json:"password"`
	// An object that contains information about the most recent in-app purchase transactions for the app.
	UnifiedReceipt UnifiedReceipt `json:"unified_receipt"`
}
