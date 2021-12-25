package apple

import (
	"encoding/json"
	"github.com/FTChinese/subscription-api/faker"
	"testing"
)

func mustReceiptResponse() VerificationResp {
	var resp VerificationResp
	if err := json.Unmarshal([]byte(faker.IAPVerificationResponse), &resp); err != nil {
		panic(err)
	}

	return resp
}

func mustParsedReceiptResponse() VerificationResp {
	resp := mustReceiptResponse()
	resp.Parse()

	return resp
}

func TestVerificationResp_ReceiptSchema(t *testing.T) {
	resp := mustParsedReceiptResponse()

	rs := resp.ReceiptSchema()

	t.Logf("%+v", rs)
}
