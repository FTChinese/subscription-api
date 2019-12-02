package apple

import (
	"github.com/guregu/null"
)

type WHLatestReceiptInfo struct {
	AppItemID int64 `json:"app_item_id"`
}

// WebHook contains the JSON data sent in the server notification from the App Store.
type WebHook struct {
	// An identifier that App Store Connect generates and the App Store uses to uniquely identify the auto-renewable subscription that the user's subscription renews
	AutoRenewAdamID int64 `json:"auto_renew_adam_id"`
	// The product identifier of the auto-renewable subscription that the user's subscription renews.
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
	LatestExpiredReceipt     string              `json:"latest_expired_receipt"`
	LatestExpiredReceiptInfo []ReceiptInfo       `json:"latest_expired_receipt_info"`
	LatestReceipt            string              `json:"latest_receipt"` // Store it as a token to verify the user's subscription status at any time.
	LatestReceiptInfo        WHLatestReceiptInfo `json:"latest_receipt_info"`
	NotificationType         NotificationType    `json:"notification_type"`
	Password                 string              `json:"password"`
	// An object that contains information about the most recent in-app purchase transactions for the app.
	UnifiedReceipt UnifiedReceipt `json:"unified_receipt"`
}

func (w WebHook) Schema(originalTransactionID string) WebHookSchema {
	return WebHookSchema{
		Environment:           w.Environment,
		OriginalTransactionID: originalTransactionID,
		AutoRenewAdamID:       w.AutoRenewAdamID,
		AutoRenewProductID:    w.AutoRenewProductID,
		AutoRenewStatus: null.NewBool(
			MustParseBoolean(w.AutoRenewStatus),
			w.AutoRenewStatus != ""),
		AutoRenewStatusChangeDateMs: MustParseInt64(w.AutoRenewStatusChangeDateMs),
		ExpirationIntent:            w.ExpirationIntent,
		NotificationType:            w.NotificationType,
		Password:                    w.Password,
		Status:                      w.UnifiedReceipt.Status,
	}
}
