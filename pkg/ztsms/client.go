package ztsms

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"github.com/FTChinese/subscription-api/lib/fetch"
	"github.com/FTChinese/subscription-api/pkg/config"
	"go.uber.org/zap"
	"strconv"
	"time"
)

type SMSSharedParams struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	Timestamp string `json:"tKey"`
}

type TemplateMessage struct {
	SMSSharedParams
	Signature  string            `json:"signature"`
	TemplateID string            `json:"tpId"`
	Records    []TemplateContent `json:"records"`
}

type TemplateContent struct {
	Mobile   string           `json:"mobile"`
	Replacer TemplateReplacer `json:"tpContent"`
}

type TemplateReplacer struct {
	Code string `json:"valid_code"`
}

// MessageResponse is the response from SMS provider's API.
// Example:
// {
// 200
// success
// 161778635408604440321
// 33337
// []
// }
type MessageResponse struct {
	Code        int               `json:"code"`
	Message     string            `json:"msg"`
	MsgID       string            `json:"msgId"`
	TemplateID  string            `json:"tpId"`
	InvalidList []TemplateContent `json:"invalidList"`
}

func (r MessageResponse) Valid() bool {
	return r.Code == 200
}

type Client struct {
	credentials config.Credentials
	logger      *zap.Logger
}

func NewClient(l *zap.Logger) Client {
	return Client{
		credentials: config.MustSMSCredentials(),
		logger:      l,
	}
}

func (c Client) hashPassword(t string) string {
	hash := md5.Sum([]byte(c.credentials.Password))
	s := hex.EncodeToString(hash[:])

	hash = md5.Sum([]byte(s + t))

	return hex.EncodeToString(hash[:])
}

func (c Client) sharedParams() SMSSharedParams {
	t := strconv.FormatInt(time.Now().Unix(), 10)

	return SMSSharedParams{
		Username:  c.credentials.Username,
		Password:  c.hashPassword(t),
		Timestamp: t,
	}
}

func (c Client) templateMessage(v Verifier) TemplateMessage {
	return TemplateMessage{
		SMSSharedParams: c.sharedParams(),
		Signature:       "【FT中文网】",
		TemplateID:      "33337",
		Records: []TemplateContent{
			{
				Mobile: v.Mobile,
				Replacer: TemplateReplacer{
					Code: v.Code,
				},
			},
		},
	}
}

func (c Client) SendVerifier(v Verifier) (MessageResponse, error) {
	defer c.logger.Sync()
	sugar := c.logger.Sugar()

	var result MessageResponse

	errs := fetch.New().
		Post("https://api.mix2.zthysms.com/v2/sendSmsTp").
		SendJSON(c.templateMessage(v)).
		EndJSON(&result)

	if errs != nil {
		sugar.Error(errs)
		return MessageResponse{}, errs[0]
	}

	sugar.Errorf("SMS response: %s", result.Message)

	if result.Valid() {
		return result, nil
	}

	return MessageResponse{}, errors.New(result.Message)
}
