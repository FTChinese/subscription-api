package apple

import (
	"sort"
	"testing"
)

func TestUnifiedReceipt_sortLatestReceiptDesc(t *testing.T) {
	ur := UnifiedReceipt{
		Environment:   EnvSandbox,
		LatestReceipt: "",
		LatestReceiptInfo: []Transaction{
			{
				ExpiresDateMs: "1532421737000",
			},
			{
				ExpiresDateMs: "1532428954000",
			},
		},
		Status:            0,
		latestTransaction: Transaction{},
	}

	ur.Parse()

	isSorted := sort.SliceIsSorted(ur.LatestReceiptInfo, func(i, j int) bool {
		return ur.LatestReceiptInfo[i].ExpiresDateMs > ur.LatestReceiptInfo[j].ExpiresDateMs
	})

	t.Logf("%t", isSorted)
}

func TestUnifiedReceipt_findPendingRenewal(t *testing.T) {
	resp := mustReceiptResponse()

	pr := resp.findPendingRenewal()

	if pr.OriginalTransactionID != "1000000595951896" {
		t.Error("wrong")
	}
}

func TestUnifiedReceipt_Subscription(t *testing.T) {
	resp := mustParsedReceiptResponse()

	sub, err := NewSubscription(resp.UnifiedReceipt)
	if err != nil {
		t.Error(err)
	}

	if sub.OriginalTransactionID == "" {
		t.Error("empty")
	}
}
