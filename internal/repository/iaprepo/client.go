package iaprepo

import (
	"github.com/FTChinese/subscription-api/pkg/apple"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/pkg/fetch"
	"io/ioutil"
	"log"
)

type Client struct {
	isSandbox  bool
	sandboxUrl string
	prodUrl    string
	password   string
}

func NewClient(sandbox bool) Client {
	return Client{
		isSandbox:  sandbox,
		sandboxUrl: "https://sandbox.itunes.apple.com/verifyReceipt",
		prodUrl:    "https://buy.itunes.apple.com/verifyReceipt",
		password:   config.MustIAPSecret(),
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

func (c Client) Verify(receipt string) ([]byte, error) {
	payload := apple.VerificationPayload{
		ReceiptData:            receipt,
		Password:               c.password,
		ExcludeOldTransactions: false,
	}

	resp, errs := fetch.NewFetch().
		Post(c.pickUrl()).
		SendJSON(payload).
		End()

	if errs != nil {
		return nil, errs[0]
	}

	return ioutil.ReadAll(resp.Body)
}
