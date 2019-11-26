package apple

// VerificationResponseBody is the response body return to verification request.
type VerificationResponseBody struct {
	Environment string `json:"environment"`  // Possible values: Sandbox | Production
	IsRetryable bool   `json:"is-retryable"` // This is only present if `Status` is not 0. 1 indicates a temporary issue; 0 indicates an unresolvable issue
	// A JSON representation of the receipt that was sent for verification
	Receipt DecodedReceipt `json:"receipt"`
	// Base-64 encoded app receipt. The latest encoded receipt, which is the same as ReceiptData in request body.
	// This is a string, not byte as specified by Apple doc. It can be decoded into bytes, which does not mean it IS byte.
	// The doc says: contains the latest encoded receipt, which is the same as the value for receipt-data in the request.
	LatestReceipt      string           `json:"latest_receipt"`
	LatestReceiptInfo  []ReceiptInfo    `json:"latest_receipt_info"`  // An array that contains all the transactions for the subscription, including the initial purchase and subsequent renewals but not including any restores.
	PendingRenewalInfo []PendingRenewal `json:"pending_renewal_info"` // each element contains the pending renewal information for each auto-renewable subscription identified by the product_id.

	// 0 if the receipt is valid, or a status code if there is an error.
	// 21000 The request to the App Store was not made using the HTTP POST request method.
	// 21001 This status code is no longer sent by the App Store.
	// 21002 The data in the receipt-data property was malformed or missing.
	// 21003 The receipt could not be authenticated.
	// 21004 The shared secret you provided does not match the shared secret on file for your account.
	// 21005 The receipt server is not currently available.
	// 21006 This receipt is valid but the subscription has expired. When this status code is returned to your server, the receipt data is also decoded and returned as part of the response. Only returned for iOS 6-style transaction receipts for auto-renewable subscriptions.
	// 21007 This receipt is from the test environment, but it was sent to the production environment for verification.
	// 21008 This receipt is from the production environment, but it was sent to the test environment for verification.
	// 21009 Internal data access error. Try again later.
	// 21010 The user account cannot be found or has been deleted.
	Status int64 `json:"status"`
}
