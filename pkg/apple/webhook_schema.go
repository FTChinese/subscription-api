package apple

import "github.com/guregu/null"

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
	ExpirationIntent            null.Int         `db:"expiration_intent"`
	NotificationType            NotificationType `db:"notification_type"`
	Password                    string           `db:"password"`
	Status                      int64            `db:"status"`
}

func NewWebHookSchema(w WebHook) WebHookSchema {
	return WebHookSchema{
		BaseTransactionSchema: w.LatestReceiptInfo.schema(
			w.Environment,
			w.LatestReceiptInfo.ExpiresDate,
		),

		AppItemID: MustParseInt64(w.LatestReceiptInfo.AppItemID),
		ItemID:    MustParseInt64(w.LatestReceiptInfo.ItemID),

		AutoRenewAdamID:    w.AutoRenewAdamID,
		AutoRenewProductID: w.AutoRenewProductID,
		AutoRenewStatus: null.NewBool(
			MustParseBoolean(w.AutoRenewStatus),
			w.AutoRenewStatus != ""),
		AutoRenewStatusChangeDateMs: MustParseInt64(w.AutoRenewStatusChangeDateMs),
		ExpirationIntent:            null.NewInt(w.ExpirationIntent, w.ExpirationIntent != 0),
		NotificationType:            w.NotificationType,
		Password:                    w.Password,
		Status:                      w.UnifiedReceipt.Status,
	}
}
