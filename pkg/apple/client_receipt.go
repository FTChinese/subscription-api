package apple

import "github.com/guregu/null"

// ClientReceipt is the receipt data from device.
// This is the value of `receipt` field from
// a verification endpoint.
// It actually contains all receipts.
// This is a very weired design since it represents a single
// transactions contains all data of transactions.
// I think the core problem is with its naming.
// Apple tries to use the word receipt naming everything.
// It does not distinguish between a receipt and a transaction.
// Its `in_app` array, `latest_receipt_info` should actually be
// taken as the history of transactions.
// They are recording of user's actions, rather than a snapshot
// of current subscription status.
type ClientReceipt struct {
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
	InAppTransactions          []Transaction `json:"in_app"`
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
