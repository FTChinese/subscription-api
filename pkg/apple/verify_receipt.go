package apple

import (
	"encoding/json"
	"github.com/parnurzeal/gorequest"
	"go.uber.org/zap"
)

var request = gorequest.New()

// VerifyReceipt sends the receipt data to the app store for verification.
// Return app store's response body.
func VerifyReceipt(payload VerificationPayload, url string) (VerificationResp, error) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	sugar := logger.Sugar()
	sugar.Infow("Verify IAP receipt",
		"trace", "apple.VerifyReceipt")

	_, body, errs := request.
		Post(url).
		Send(payload).End()

	if errs != nil {
		sugar.Error(errs)
		return VerificationResp{}, errs[0]
	}

	sugar.Infof("IAP verification response: %s", body)

	var resp VerificationResp
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		sugar.Error(err)
		return resp, err
	}

	return resp, nil
}
