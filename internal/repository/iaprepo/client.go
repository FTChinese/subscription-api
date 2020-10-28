package iaprepo

import (
	"encoding/json"
	"github.com/FTChinese/subscription-api/pkg/apple"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/pkg/fetch"
	"go.uber.org/zap"
	"log"
)

type Client struct {
	isSandbox  bool
	sandboxUrl string
	prodUrl    string
	password   string
	logger     *zap.Logger
}

func NewClient(sandbox bool, logger *zap.Logger) Client {
	return Client{
		isSandbox:  sandbox,
		sandboxUrl: "https://sandbox.itunes.apple.com/verifyReceipt",
		prodUrl:    "https://buy.itunes.apple.com/verifyReceipt",
		password:   config.MustIAPSecret(),
		logger:     logger,
	}
}

func (c Client) pickUrl() string {
	if c.isSandbox {
		log.Print("Using IAP sandbox url")
		return c.sandboxUrl
	}

	log.Print("Using IAP production url")
	return c.prodUrl
}

func (c Client) Verify(receipt string) (apple.VerificationResp, error) {
	defer c.logger.Sync()
	sugar := c.logger.Sugar()

	payload := apple.VerificationPayload{
		ReceiptData:            receipt,
		Password:               c.password,
		ExcludeOldTransactions: false,
	}

	resp, b, errs := fetch.New().
		Post(c.pickUrl()).
		SendJSON(payload).
		EndRaw()

	sugar.Infof("App store response status code %d", resp.StatusCode)
	sugar.Infof("App store verification raw content %s", resp.StatusCode, b)

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
