package apple

// VerificationRequestBody contains the the JSON contents
// you submit with the request to the App Store when verifying receipt.
// The ReceiptData is encrypted and the Password is used to decrypt it.
type VerificationRequestBody struct {
	ReceiptData            string `json:"receipt-data"`
	Password               string `json:"password"`
	ExcludeOldTransactions bool   `json:"exclude-old-transactions"`
}
