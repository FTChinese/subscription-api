package poll

import (
	"encoding/json"
	"errors"

	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/lib/fetch"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/tidwall/gjson"
)

type APIClient struct {
	key     string
	baseURL string
}

func NewAPIClient(prod bool) APIClient {
	return APIClient{
		key:     config.MustLoadPollingKey().Pick(prod),
		baseURL: config.MustSubsAPIv1BaseURL().Pick(prod),
	}
}

func (c APIClient) GetReceipt(origTxID string) (string, error) {
	url := c.baseURL + "/apple/receipt/" + origTxID + "?fs=true"

	resp, b, errs := fetch.New().Get(url).SetBearerAuth(c.key).EndBytes()
	if errs != nil {
		return "", errs[0]
	}

	if resp.StatusCode >= 400 {
		var respErr render.ResponseError
		if err := json.Unmarshal(b, &respErr); err != nil {
			return "", err
		}

		return "", &respErr
	}

	result := gjson.GetBytes(b, "receipt")

	if !result.Exists() {
		return "", errors.New("receipt not found from subscription api")
	}

	return result.String(), nil
}
