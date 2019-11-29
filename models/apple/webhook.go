package apple

import "github.com/guregu/null"

type NotificationType string

const (
	NotificationTypeCancel                 NotificationType = "CANCEL"                    // Indicates that either Apple customer support canceled the subscription or the user upgraded their subscription
	NotificationTypeDidChangeRenewalPref                    = "DID_CHANGE_RENEWAL_PREF"   // DID_CHANGE_RENEWAL_PREF
	NotificationTypeDidChangeRenewalStatus                  = "DID_CHANGE_RENEWAL_STATUS" // Indicates a change in the subscription renewal status.
	NotificationTypeDidFailToRenew                          = "DID_FAIL_TO_RENEW"         // Indicates a subscription that failed to renew due to a billing issue.
	NotificationTypeDidRecover                              = "DID_RECOVER"               // Indicates successful automatic renewal of an expired subscription that failed to renew in the past.
	NotificationTypeInitialBuy                              = "INITIAL_BUY"               // Occurs at the initial purchase of the subscription
	NotificationTypeInteractiveRenewal                      = "INTERACTIVE_RENEWAL"       // Indicates the customer renewed a subscription interactively, either by using your app’s interface, or on the App Store in the account's Subscriptions settings.
	NotificationTypeRenewal                                 = "RENEWAL"                   // Indicates successful automatic renewal of an expired subscription that failed to renew in the past.
)

// UnifiedReceipt is an object that contains information about the most recent in-app purchase transactions for the app.
type UnifiedReceipt struct {
	Environment        Environment      `json:"environment"`
	LatestReceipt      string           `json:"latest_receipt"`
	LatestReceiptInfo  []ReceiptInfo    `json:"latest_receipt_info"`
	PendingRenewalInfo []PendingRenewal `json:"pending_renewal_info"`
	Status             int64            `json:"status"`
}

// WebHook contains the JSON data sent in the server notification from the App Store.
type WebHook struct {
	AutoRenewAdamID    int64  `json:"auto_renew_adam_id"`
	AutoRenewProductID string `json:"auto_renew_product_id"`
	// The current renewal status for an auto-renewable subscription product.
	// Note that these values are different from those of the auto_renew_status in the receipt.
	// true | false
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
	ExpirationIntent null.String `json:"expiration_intent"`

	// The latest Base64-encoded transaction receipt
	LatestExpiredReceipt     string           `json:"latest_expired_receipt"`
	LatestExpiredReceiptInfo []ReceiptInfo    `json:"latest_expired_receipt_info"`
	LatestReceipt            string           `json:"latest_receipt"` // Store it as a token to verify the user's subscription status at any time.
	LatestReceiptInfo        []ReceiptInfo    `json:"latest_receipt_info"`
	NotificationType         NotificationType `json:"notification_type"`
	Password                 string           `json:"password"`
	// An object that contains information about the most recent in-app purchase transactions for the app.
	UnifiedReceipt UnifiedReceipt `json:"unified_receipt"`
}
