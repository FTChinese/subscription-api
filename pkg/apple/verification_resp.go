package apple

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/guregu/null"
	"strconv"
)

const bundleID = "com.ft.ftchinese.mobile"

// VerificationResp is the response body return to verification request.
// Per the doc:
// For auto-renewable subscription items, parse the response to get information about the currently active subscription period.
// When you validate the receipt for a subscription, `latest_receipt` contains the latest encoded receipt,
// which is the same as the value for `receipt-data` in the request, and `latest_receipt_info` contains
// all the transactions for the subscription, including the initial purchase and subsequent renewals but not including any restores.
//
// You can use these values to check whether an auto-renewable subscription has expired.
// Use these values along with the `expiration_intent` subscription field to get the reason for expiration
//
// See https://developer.apple.com/documentation/appstorereceipts/requestbody
type VerificationResp struct {
	UnifiedReceipt
	// This is only present if `Status` is not 0. 1 indicates a temporary issue; 0 indicates an unresolvable issue
	// Only applicable to status codes 21100-21199.
	IsRetryable bool `json:"is-retryable"`
	// The decoded version of the encoded receipt you send to app store for verification.
	// It is the decoded version of VerificationPayload.ReceiptData
	Receipt ClientReceipt `json:"receipt"`
}

// Validate ensures the response from Apple API is correct.
// Checks Status and BundleID.
func (v *VerificationResp) Validate() *render.ValidationError {

	// Status above 0 is error.
	if v.Status != 0 {
		return &render.ValidationError{
			Message: getStatusMessage(v.Status),
			Field:   "status",
			Code:    render.InvalidCode(strconv.FormatInt(v.Status, 10)),
		}
	}

	if v.Receipt.BundleID != bundleID {
		return &render.ValidationError{
			Message: "Bundle id mismatched",
			Field:   "bundle_id",
			Code:    render.CodeInvalid,
		}
	}

	// LatestReceiptInfo should not be empty.
	if !v.IsSubscribed() {
		return &render.ValidationError{
			Message: "No subscription found",
			Field:   "latest_receipt_info",
			Code:    render.CodeMissingField,
		}
	}

	return nil
}

// ReceiptSchema is used to save the decoded ClientReceipt received in a verification response.
// Every verification request will create this record.
// You must call UnifiedReceipt.Parse before building it.
func (v *VerificationResp) ReceiptSchema() VerifiedReceiptSchema {

	receiptType, _ := ParseReceiptType(v.Receipt.ReceiptType)

	return VerifiedReceiptSchema{
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
