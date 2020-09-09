package apple

import (
	"github.com/stretchr/testify/assert"
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

	assert.True(t, isSorted)

	assert.Equal(t, ur.latestTransaction.ExpiresDateMs, "1532428954000")
}

func TestUnifiedReceipt_findPendingRenewal(t *testing.T) {
	resp := mustReceiptResponse()

	pr := resp.findPendingRenewal()

	assert.Equal(t, pr.OriginalTransactionID, "1000000595951896")
}

func TestUnifiedReceipt_Subscription(t *testing.T) {
	resp := mustParsedReceiptResponse()

	sub, err := resp.Subscription()
	if err != nil {
		t.Error(err)
	}

	assert.NotEmpty(t, sub.OriginalTransactionID)
}
