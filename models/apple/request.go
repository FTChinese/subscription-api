package apple

import (
	"github.com/FTChinese/go-rest/view"
	"strings"
)

// VerificationRequestBody contains the the JSON contents
// you submit with the request to the App Store when verifying receipt.
// The ReceiptData is encrypted and the Password is used to decrypt it.
type VerificationRequestBody struct {
	ReceiptData            string `json:"receipt-data"`
	Password               string `json:"password"`
	ExcludeOldTransactions bool   `json:"exclude-old-transactions"`
}

func (v VerificationRequestBody) Validate() *view.Reason {
	if strings.TrimSpace(v.ReceiptData) == "" {
		r := view.NewReason()
		r.Field = "receipt-data"
		r.Code = view.CodeMissingField
		return r
	}

	return nil
}

func (v VerificationRequestBody) ReceiptToken(e Environment, originalTransactionID string) ReceiptToken {
	return ReceiptToken{
		BaseSchema: BaseSchema{
			Environment:           e,
			OriginalTransactionID: originalTransactionID,
		},
		LatestReceipt: v.ReceiptData,
	}
}
