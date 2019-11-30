package apple

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/guregu/null"
	"github.com/sirupsen/logrus"
	plan2 "gitlab.com/ftchinese/subscription-api/models/plan"
	"gitlab.com/ftchinese/subscription-api/models/reader"
	"sort"
	"time"
)

const bundleID = "com.ft.ftchinese.mobile"

// VerificationResponseBody is the response body return to verification request.
type VerificationResponseBody struct {
	Environment Environment `json:"environment"` // Possible values: Sandbox | Production
	// This is only present if `Status` is not 0. 1 indicates a temporary issue; 0 indicates an unresolvable issue
	// Only applicable to status codes 21100-21199.
	IsRetryable bool `json:"is-retryable"`
	// A JSON representation of the receipt that was sent for verification
	Receipt DecodedReceipt `json:"receipt"`
	// Base-64 encoded app receipt. The latest encoded receipt, which is the same as ReceiptData in request body.
	// This is a string, not byte as specified by Apple doc. It can be decoded into bytes, which does not mean it IS byte.
	// The doc says: contains the latest encoded receipt, which is the same as the value for receipt-data in the request.
	LatestReceipt string `json:"latest_receipt"`
	// An array that contains all the transactions for the subscription, including the initial purchase and subsequent renewals but not including any restores.
	// Use this to get the status of the most recent renewal
	LatestReceiptInfo  []ReceiptInfo    `json:"latest_receipt_info"`
	PendingRenewalInfo []PendingRenewal `json:"pending_renewal_info"` // each element contains the pending renewal information for each auto-renewable subscription identified by the product_id.

	// The value for status is 0 if the receipt is valid.
	// The status code reflects the status of the app receipt as a whole.
	// For example, if you send a valid app receipt that contains an expired subscription, the response is 0 because the receipt is valid.
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
	// 21100-21199 are various internal data access errors.
	Status int64 `json:"status"`
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

	if v.LatestReceiptInfo == nil || len(v.LatestReceiptInfo) == 0 {
		logger.Infof("latest_receipt_info field is either nil or empty")
		return false
	}

	return true
}

func (v *VerificationResponseBody) Failure(receiptData string) VerificationFailed {
	m := getStatusMessage(v.Status)

	return VerificationFailed{
		Environment: v.Environment,
		Status:      v.Status,
		Message:     null.NewString(m, m != ""),
		ReceiptData: receiptData,
	}
}

func (v *VerificationResponseBody) SortLatestReceiptsDesc() {
	sort.SliceStable(v.LatestReceiptInfo, func(i, j int) bool {
		return v.LatestReceiptInfo[i].ExpiresDateMs > v.LatestReceiptInfo[j].ExpiresDateMs
	})
}

func (v *VerificationResponseBody) FindLatestReceipt() ReceiptInfo {
	v.SortLatestReceiptsDesc()

	l := len(v.LatestReceiptInfo)
	nowUnix := time.Now().Unix()

	// Find the first one whose expires date is greater than now,
	// and not cancelled,
	// and is a valid product.
	i := sort.Search(l, func(i int) bool {
		r := v.LatestReceiptInfo[i]

		return r.ExpiresUnix() >= nowUnix && !r.IsCancelled() && r.IsValidProduct()
	})

	if i < l {
		return v.LatestReceiptInfo[i]
	}

	return v.LatestReceiptInfo[0]
}

func (v *VerificationResponseBody) SessionSchema(r ReceiptInfo) VerificationSessionSchema {

	receiptType, _ := ParseReceiptType(v.Receipt.ReceiptType)

	return VerificationSessionSchema{
		Status:                    v.Status,
		Environment:               v.Environment,
		OriginalTransactionID:     r.OriginalTransactionID,
		TransactionID:             r.TransactionID,
		ReceiptType:               receiptType,
		AppItemID:                 v.Receipt.AppItemID,
		BundleID:                  v.Receipt.BundleID,
		ApplicationVersion:        v.Receipt.ApplicationVersion,
		DownloadID:                v.Receipt.DownloadID,
		VersionExternalIdentifier: v.Receipt.VersionExternalIdentifier,
		ReceiptCreationDateMs:     MustParseInt64(v.Receipt.ReceiptCreationDateMs),
		RequestDateMs:             MustParseInt64(v.Receipt.RequestDateMs),
		OriginalPurchaseDateMs:    MustParseInt64(v.Receipt.OriginalPurchaseDateMs),
		ExpirationDateMs: null.NewInt(
			MustParseInt64(v.Receipt.ExpirationDateMs.String),
			v.Receipt.ExpirationDateMs.Valid,
		),
		OriginalApplicationVersion: v.Receipt.OriginalApplicationVersion,
		PreorderDateMs: null.NewInt(
			MustParseInt64(v.Receipt.PreorderDateMs.String),
			v.Receipt.PreorderDateMs.Valid,
		),
	}
}

func (v *VerificationResponseBody) ReceiptToken(originalTransactionID string) ReceiptToken {
	return ReceiptToken{
		Environment:           v.Environment,
		OriginalTransactionID: originalTransactionID,
		LatestReceipt:         v.LatestReceipt,
	}
}

func (v *VerificationResponseBody) findPendingRenewal(r ReceiptInfo) PendingRenewal {

	for _, v := range v.PendingRenewalInfo {
		if v.OriginalTransactionID == r.OriginalTransactionID && (v.ProductID == r.ProductID) {
			return v
		}
	}

	return PendingRenewal{}
}

// Subscription builds a subscription for a user based on
// the receipt information available.
func (v *VerificationResponseBody) Subscription(ids reader.MemberID, r ReceiptInfo) Subscription {

	p := v.findPendingRenewal(r)

	plan, _ := plan2.FindPlanForApple(r.ProductID)
	pd := MustParseInt64(r.PurchaseDateMs) / 1000

	return Subscription{
		Environment:           v.Environment,
		OriginalTransactionID: r.OriginalTransactionID,
		LastTransactionID:     r.TransactionID,
		ProductID:             r.ProductID,
		PurchaseDateUTC: chrono.TimeFrom(
			time.Unix(pd, 0),
		),
		ExpiresDateUTC: chrono.TimeFrom(
			time.Unix(r.ExpiresUnix(), 0),
		),
		MemberID:    ids,
		Tier:        plan.Tier,
		Cycle:       plan.Cycle,
		AutoRenewal: p.IsAutoRenew(),
	}
}
