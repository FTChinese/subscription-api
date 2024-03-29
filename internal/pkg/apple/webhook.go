package apple

// WebHook contains the JSON data sent in the server notification from the App Store.
// It seems the top level data is a mix of ClientReceipt
// and PendingRenewal.
type WebHook struct {
	// An identifier that App Store Connect generates and the App Store uses to uniquely identify the auto-renewable subscription that the user's subscription renews
	AutoRenewAdamID int64 `json:"auto_renew_adam_id"`
	// The price identifier of the auto-renewable subscription that the user's subscription renews.
	// Also present in PendingRenewal
	AutoRenewProductID string `json:"auto_renew_product_id"`
	// The current renewal status for an auto-renewable subscription price.
	// Note that these values are different from those of the auto_renew_status in the receipt.
	// true | false
	// Also present in PendingRenewal
	AutoRenewStatus string `json:"auto_renew_status"`

	// The time at which the renewal status for an auto-renewable subscription was turned on or off,
	AutoRenewStatusChangeDate    string `json:"auto_renew_status_change_date"`
	AutoRenewStatusChangeDateMs  string `json:"auto_renew_status_change_date_ms"`
	AutoRenewStatusChangeDatePST string `json:"auto_renew_status_change_date_pst"`

	BundleID      string `json:"bid"` //  the app bundle ID.
	BundleVersion string `json:"bvrs"`

	// The environment for which the receipt was generated.
	//Possible values: Sandbox, PROD
	Environment Environment `json:"environment"` // the app bundle version.

	// The reason a subscription expired.
	// This field is only present for an expired auto-renewable subscription.
	// 1 The customer voluntarily canceled their subscription.
	// 2 Billing error; for example, the customer's payment information was no longer valid.
	// 3 The customer did not agree to a recent price increase.
	// 4 The price was not available for purchase at the time of renewal.
	// 5 Unknown error.
	// Also found in PendingRenewal.
	ExpirationIntent int64 `json:"expiration_intent"`
	//LatestReceiptInfo LatestTransaction `json:"latest_receipt_info"` // Deprecated. Use the unified_receipt.Latest_receipt_info
	NotificationType NotificationType `json:"notification_type"`
	Password         string           `json:"password"`
	// An object that contains information about the most recent in-app purchase transactions for the app.
	UnifiedReceipt UnifiedReceipt `json:"unified_receipt"`
}
