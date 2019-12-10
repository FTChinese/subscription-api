package apple

import (
	"github.com/guregu/null"
	"github.com/sirupsen/logrus"
)

const bundleID = "com.ft.ftchinese.mobile"

// VerificationResponseBody is the response body return to verification request.
type VerificationResponseBody struct {
	UnifiedReceipt
	// This is only present if `Status` is not 0. 1 indicates a temporary issue; 0 indicates an unresolvable issue
	// Only applicable to status codes 21100-21199.
	IsRetryable bool `json:"is-retryable"`
	// A JSON representation of the receipt that was sent for verification
	Receipt ClientReceipt `json:"receipt"`
}

func (v *VerificationResponseBody) GetStatusMessage() string {
	if v.Status >= 21100 && v.Status <= 21199 {
		return "Internal data access errors"
	}

	return statusMessage[v.Status]
}

// Validate ensures the response from Apple API is correct.
func (v *VerificationResponseBody) Validate() bool {
	logger := logrus.WithField("project", "subscription-api").
		WithField("package", "models.apple").
		WithField("trace", "VerificationResponseBody")

	if v.Status != 0 {
		logger.Infof("Expected response status 0, got %d: %s", v.Status, getStatusMessage(v.Status))
		return false
	}

	if v.Receipt.BundleID != bundleID {
		logger.Infof("Bundle ID does not match, got %s", v.Receipt.BundleID)
		return false
	}

	if v.LatestTransactions == nil || len(v.LatestTransactions) == 0 {
		logger.Infof("latest_receipt_info field is either nil or empty")
		return false
	}

	return true
}

// SessionSchema builds the schema by merging the root elements
// of verification, the Receipt top fields and the latest transaction ids.
func (v *VerificationResponseBody) SessionSchema() VerificationSessionSchema {

	receiptType, _ := ParseReceiptType(v.Receipt.ReceiptType)

	return VerificationSessionSchema{
		BaseSchema: BaseSchema{
			Environment:           v.Environment,
			OriginalTransactionID: v.latestTransaction.OriginalTransactionID,
		},
		TransactionID: v.latestTransaction.TransactionID,

		AppItemID:          v.Receipt.AppItemID,
		ApplicationVersion: v.Receipt.ApplicationVersion,
		BundleID:           v.Receipt.BundleID,
		DownloadID:         v.Receipt.DownloadID,
		ExpirationDateMs: null.NewInt(
			MustParseInt64(v.Receipt.ExpirationDateMs.String),
			v.Receipt.ExpirationDateMs.Valid,
		),
		OriginalPurchaseDateMs:     MustParseInt64(v.Receipt.OriginalPurchaseDateMs),
		OriginalApplicationVersion: v.Receipt.OriginalApplicationVersion,
		PreorderDateMs: null.NewInt(
			MustParseInt64(v.Receipt.PreorderDateMs.String),
			v.Receipt.PreorderDateMs.Valid,
		),
		ReceiptCreationDateMs:     MustParseInt64(v.Receipt.ReceiptCreationDateMs),
		ReceiptType:               receiptType,
		RequestDateMs:             MustParseInt64(v.Receipt.RequestDateMs),
		Status:                    v.Status,
		VersionExternalIdentifier: v.Receipt.VersionExternalIdentifier,
	}
}
