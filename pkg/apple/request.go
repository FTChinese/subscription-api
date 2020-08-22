package apple

import (
	"encoding/json"
	"github.com/parnurzeal/gorequest"
)

var request = gorequest.New()

// VerifyReceipt sends the receipt data to the app store for verification.
// Return app store's response body.
func VerifyReceipt(payload VerificationPayload, url string) (VerificationResp, error) {

	_, body, errs := request.
		Post(url).
		Send(payload).End()

	if errs != nil {
		return VerificationResp{}, errs[0]
	}

	// TODO: add logger

	var resp VerificationResp
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		return resp, err
	}

	return resp, nil
}
