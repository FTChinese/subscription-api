package apple

import (
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/config"
	"testing"
)

func TestVerifyReceipt(t *testing.T) {
	payload := VerificationPayload{
		ReceiptData:            faker.IAPReceipt,
		Password:               config.MustGetIAPSecret(),
		ExcludeOldTransactions: false,
	}

	cfg := config.NewBuildConfig(false, false)
	resp, err := VerifyReceipt(payload, cfg.GetIAPVerificationURL())

	if err != nil {
		t.Error(err)
	}

	t.Logf("%s", resp.LatestToken)
}
