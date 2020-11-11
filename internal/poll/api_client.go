package poll

import (
	"encoding/json"
	"errors"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/pkg/fetch"
	"github.com/tidwall/gjson"
)

type APIClient struct {
	key     string
	baseURL string
}

func NewAPIClient(prod bool) APIClient {
	return APIClient{
		key:     config.MustAPIKey().Pick(prod),
		baseURL: config.MustAPIBaseURL().Pick(prod),
	}
}

func (c APIClient) GetReceipt(origTxID string) (string, error) {
	url := c.baseURL + "/apple/receipt/" + origTxID

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
