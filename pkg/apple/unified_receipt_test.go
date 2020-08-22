package apple

import (
	"sort"
	"testing"
)

func TestUnifiedReceipt_sortLatestReceiptDesc(t *testing.T) {
	ur := UnifiedReceipt{
		Environment: EnvSandbox,
		LatestToken: "",
		LatestTransactions: []Transaction{
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

	isSorted := sort.SliceIsSorted(ur.LatestTransactions, func(i, j int) bool {
		return ur.LatestTransactions[i].ExpiresDateMs > ur.LatestTransactions[j].ExpiresDateMs
	})

	if !isSorted {
		t.Error("Not sorted")
	}

	if ur.latestTransaction.ExpiresDateMs != "1532428954000" {
		t.Error("latest not found")
	}
}

func TestUnifiedReceipt_findPendingRenewal(t *testing.T) {
	resp := mustReceiptResponse()

	pr := resp.findPendingRenewal()

	if pr.OriginalTransactionID != "1000000595951896" {
		t.Error("not found")
	}
}
