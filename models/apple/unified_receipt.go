package apple

import (
	"github.com/FTChinese/go-rest/chrono"
	"gitlab.com/ftchinese/subscription-api/models/reader"
	"sort"
	"time"
)

// UnifiedReceipt is the payload apple send in a notification,
// or the base response to verify a receipt.
type UnifiedReceipt struct {
	// The environment for which the receipt was generated.
	// Possible values: Sandbox, Production
	Environment Environment `json:"environment"`
	// The latest Base64-encoded app receipt.
	// For verification, this is the latest encoded receipt.
	// This is a string, not byte as specified by Apple doc. It can be decoded into bytes, which does not mean it IS byte.
	LatestReceipt string `json:"latest_receipt"`
	// An array that contains the latest 100 in-app purchase transactions of the decoded value in latest_receipt.
	//  For verification, an array that contains all the transactions for the subscription, including the initial purchase and subsequent renewals but not including any restores.
	// Use this to get the status of the most recent renewal
	LatestReceiptInfo []ReceiptInfo `json:"latest_receipt_info"`
	// An array where each element contains the pending renewal information for each auto-renewable subscription identified in product_id.
	// each element contains the pending renewal information for each auto-renewable subscription identified by the product_id.
	PendingRenewalInfo []PendingRenewal `json:"pending_renewal_info"`

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
	// For notification the status code, where 0 indicates that the notification is valid.
	Status int64 `json:"status"`
}

func (u *UnifiedReceipt) Validate() bool {
	if u.LatestReceiptInfo == nil || len(u.LatestReceiptInfo) == 0 {
		return false
	}

	return true
}

func (u *UnifiedReceipt) SortLatestReceiptsDesc() {
	sort.SliceStable(u.LatestReceiptInfo, func(i, j int) bool {
		return u.LatestReceiptInfo[i].ExpiresDateMs > u.LatestReceiptInfo[j].ExpiresDateMs
	})
}

func (u *UnifiedReceipt) FindLatestReceipt() ReceiptInfo {
	u.SortLatestReceiptsDesc()

	l := len(u.LatestReceiptInfo)

	nowUnix := time.Now().Unix()

	i := sort.Search(l, func(i int) bool {
		r := u.LatestReceiptInfo[i]

		return r.ExpiresUnix() >= nowUnix && !r.IsCancelled() && r.IsValidProduct()
	})

	if i < l {
		return u.LatestReceiptInfo[i]
	}

	return u.LatestReceiptInfo[0]
}

func (u *UnifiedReceipt) ReceiptToken(originalTransactionID string) ReceiptToken {
	return ReceiptToken{
		Environment:           u.Environment,
		OriginalTransactionID: originalTransactionID,
		LatestReceipt:         u.LatestReceipt,
	}
}

// findPendingRenewal loops over the PendingRenewalInfo array
// to find the element that matches the latest receipt.
// Returns a zero instance if not found.
// The zero value if valid since we're only interested in
// the auto renew field which should default to false.
func (u *UnifiedReceipt) findPendingRenewal(r ReceiptInfo) PendingRenewal {
	if u.PendingRenewalInfo == nil {
		return PendingRenewal{}
	}

	for _, v := range u.PendingRenewalInfo {
		if v.OriginalTransactionID == r.OriginalTransactionID && (v.ProductID == r.ProductID) {
			return v
		}
	}

	return PendingRenewal{}
}

// Subscription builds a subscription for a user based on
// the receipt information available.
func (u *UnifiedReceipt) Subscription(ids reader.MemberID, r ReceiptInfo) Subscription {
	pendingRenewal := u.findPendingRenewal(r)

	p := r.FindPlan()

	return Subscription{
		Environment:           u.Environment,
		OriginalTransactionID: r.OriginalTransactionID,
		LastTransactionID:     r.TransactionID,
		ProductID:             r.ProductID,
		PurchaseDateUTC: chrono.TimeFrom(
			time.Unix(r.PurchaseDateUnix(), 0),
		),
		ExpiresDateUTC: chrono.TimeFrom(
			time.Unix(r.ExpiresUnix(), 0),
		),
		MemberID:    ids,
		Tier:        p.Tier,
		Cycle:       p.Cycle,
		AutoRenewal: pendingRenewal.IsAutoRenew(),
	}
}
