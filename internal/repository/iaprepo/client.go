package iaprepo

import (
	"encoding/json"
	"github.com/FTChinese/subscription-api/lib/fetch"
	"github.com/FTChinese/subscription-api/pkg/apple"
	"github.com/FTChinese/subscription-api/pkg/config"
	"go.uber.org/zap"
	"log"
)

func verifyURL(sandbox bool) string {
	if sandbox {
		log.Print("Using IAP sandbox url")
		return "https://sandbox.itunes.apple.com/verifyReceipt"
	}

	log.Print("Using IAP production url")
	return "https://buy.itunes.apple.com/verifyReceipt"
}

type Client struct {
	password string
	logger   *zap.Logger
}

func NewClient(logger *zap.Logger) Client {
	return Client{
		password: config.MustIAPSecret(),
		logger:   logger,
	}
}

// Verify the passed in receipt against App Store.
// `sandbox` determines which endpoint to use.
// For subscription-api, sandbox is determined by cli args upon startup. It will never change after server started.
// However, when polling App Store, sandbox is dynamically determined
// by environment value retrieved from DB.
func (c Client) Verify(receipt string, sandbox bool) (apple.VerificationResp, error) {
	defer c.logger.Sync()
	sugar := c.logger.Sugar()

	payload := apple.VerificationPayload{
		ReceiptData:            receipt,
		Password:               c.password,
		ExcludeOldTransactions: false,
	}

	resp, b, errs := fetch.New().
		Post(verifyURL(sandbox)).
		SendJSON(payload).
		EndBytes()

	sugar.Infof("App store response status code %d", resp.StatusCode)
	sugar.Infof("App store verification raw content %s", b)

	if errs != nil {
		return apple.VerificationResp{}, errs[0]
	}

	var vrfResult apple.VerificationResp
	if err := json.Unmarshal(b, &vrfResult); err != nil {
		return apple.VerificationResp{}, err
	}

	sugar.Infof("Environment %s, is retryable %t, status %d", vrfResult.Environment, vrfResult.IsRetryable, vrfResult.Status)

	return vrfResult, nil
}

// VerifyAndValidate verifies an receipt and validate if the response is valid.
// The return error might be an instance of render.ValidationError.
func (c Client) VerifyAndValidate(receipt string, sandbox bool) (apple.VerificationResp, error) {
	resp, err := c.Verify(receipt, sandbox)
	if err != nil {
		return apple.VerificationResp{}, err
	}

	if err := resp.Validate(); err != nil {
		return apple.VerificationResp{}, err
	}

	resp.Parse()

	return resp, nil
}
