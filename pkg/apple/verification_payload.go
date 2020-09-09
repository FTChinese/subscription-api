package apple

// VerificationPayload contains the the JSON contents
// you submit with the request to the App Store when verifying receipt.
// The ReceiptData is encrypted and the Password is used to decrypt it.
// Per the doc https://developer.apple.com/documentation/storekit/in-app_purchase/validating_receipts_with_the_app_store:
// On your server, create a JSON object with the
// * `receipt-data`,
// * `password` (if the receipt contains an auto-renewable subscription), and
// * `exclude-old-transactions` keys.
// Submit this JSON object as the payload of an HTTP POST request.
// Use the test environment URL https://sandbox.itunes.apple.com/verifyReceipt.
// Use the production URL https://buy.itunes.apple.com/verifyReceipt.
// See https://developer.apple.com/documentation/appstorereceipts/requestbody
type VerificationPayload struct {
	ReceiptData            string `json:"receipt-data"`             // Required. The Base64 encoded receipt data.
	Password               string `json:"password"`                 // Required. Your app's shared secret (a hexadecimal string).
	ExcludeOldTransactions bool   `json:"exclude-old-transactions"` // Set this value to true for the response to include only the latest renewal transaction for any subscriptions. Applicable only to auto-renewable subscriptions.
}
