package apple

type WebHookBody struct {
	AutoRenewAdamID              int64         `json:"auto_renew_adam_id"` // uniquely identify the auto-renewable subscription that the user's subscription renews
	AutoRenewProductID           string        `json:"auto_renew_product_id"`
	AutoRenewStatus              string        `json:"auto_renew_status"`
	AutoRenewStatusChangeDate    string        `json:"auto_renew_status_change_date"`
	AutoRenewStatusChangeDateMs  string        `json:"auto_renew_status_change_date_ms"`
	AutoRenewStatusChangeDatePST string        `json:"auto_renew_status_change_date_pst"`
	Environment                  string        `json:"environment"` // Sandbox | PROD
	ExpirationIntent             int64         `json:"expiration_intent"`
	LatestExpiredReceipt         byte          `json:"latest_expired_receipt"`
	LatestReceiptInfo            []ReceiptInfo `json:"latest_receipt_info"`
	NotificationType             string        `json:"notification_type"` // CANCEL | DID_CHANGE_RENEWAL_PREF | DID_CHANGE_RENEWAL_STATUS | DID_FAIL_TO_RENEWAL | DID_RECOVER | INITIAL_BUY | INTERACTIVE_RENEWAL | RENEWAL
}
