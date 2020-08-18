package apple

import (
	"github.com/FTChinese/go-rest/chrono"
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
	LatestToken string `json:"latest_receipt"`
	// An array that contains the latest 100 in-app purchase transactions of the decoded value in latest_receipt.
	//  For verification, an array that contains all the transactions for the subscription, including the initial purchase and subsequent renewals but not including any restores.
	// Use this to get the status of the most recent renewal
	LatestTransactions []Transaction `json:"latest_receipt_info"`
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
	Status            int64       `json:"status"`
	latestTransaction Transaction // hold the latest transaction sorted from LatestTransactions array.
}

func (u *UnifiedReceipt) Validate() bool {
	if u.LatestTransactions == nil || len(u.LatestTransactions) == 0 {
		return false
	}

	return true
}

func (u *UnifiedReceipt) sortLatestReceiptsDesc() {
	sort.SliceStable(u.LatestTransactions, func(i, j int) bool {
		return u.LatestTransactions[i].ExpiresDateMs > u.LatestTransactions[j].ExpiresDateMs
	})
}

func (u *UnifiedReceipt) findLatestTransaction() Transaction {
	u.sortLatestReceiptsDesc()

	l := len(u.LatestTransactions)

	nowUnix := time.Now().Unix()

	// It is questionable to check the if it is cancelled.
	// I think the most recent receipt is always reliable.
	// Those record should be generated linearly.
	// If you cancelled a subscription, later re-subscribed,
	// the cancelled receipt cannot be one the top.
	// It would be unreasonable if the the most recent one
	// is cancelled while a previous one is still valid.
	// If the above is true, it only indicates the design of
	// your product and pricing system have problems.
	i := sort.Search(l, func(i int) bool {
		r := u.LatestTransactions[i]

		return r.ExpiresUnix() >= nowUnix && r.IsValidProduct()
	})

	if i < l {
		return u.LatestTransactions[i]
	}

	return u.LatestTransactions[0]
}

func (u *UnifiedReceipt) Parse() {
	u.latestTransaction = u.findLatestTransaction()
}

func (u *UnifiedReceipt) ReceiptToken() ReceiptToken {
	return ReceiptToken{
		BaseSchema: BaseSchema{
			Environment:           u.Environment,
			OriginalTransactionID: u.latestTransaction.OriginalTransactionID,
		},
		LatestReceipt: u.LatestToken,
	}
}

// findPendingRenewal loops over the PendingRenewalInfo array
// to find the element that matches the latest receipt.
// Returns a zero instance if not found.
// The zero value if valid since we're only interested in
// the auto renew field which should default to false.
func (u *UnifiedReceipt) findPendingRenewal() PendingRenewal {
	if u.PendingRenewalInfo == nil {
		return PendingRenewal{}
	}

	for _, v := range u.PendingRenewalInfo {
		if v.OriginalTransactionID == u.latestTransaction.OriginalTransactionID && (v.ProductID == u.latestTransaction.ProductID) {
			return v
		}
	}

	return PendingRenewal{}
}

// Subscription builds a subscription for a user based on
// the receipt information available.
// TODO: What if the Transaction is a cancelled one?
func (u *UnifiedReceipt) Subscription() Subscription {
	pendingRenewal := u.findPendingRenewal()

	autoRenew := pendingRenewal.IsAutoRenew()
	if u.latestTransaction.IsCancelled() {
		autoRenew = false
	}

	p := u.latestTransaction.FindPlan()

	return Subscription{
		Environment:           u.Environment,
		OriginalTransactionID: u.latestTransaction.OriginalTransactionID,
		LastTransactionID:     u.latestTransaction.TransactionID,
		ProductID:             u.latestTransaction.ProductID,
		PurchaseDateUTC: chrono.TimeFrom(
			time.Unix(u.latestTransaction.PurchaseDateUnix(), 0),
		),
		ExpiresDateUTC: chrono.TimeFrom(
			time.Unix(u.latestTransaction.ExpiresUnix(), 0),
		),
		BasePlan:    p.BasePlan,
		AutoRenewal: autoRenew,
	}
}
