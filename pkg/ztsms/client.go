package ztsms

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"github.com/FTChinese/subscription-api/lib/fetch"
	"github.com/FTChinese/subscription-api/pkg/config"
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
	TemplateID int64             `json:"tpId"`
	Records    []TemplateContent `json:"records"`
}

type TemplateContent struct {
	Mobile   string           `json:"mobile"`
	Replacer TemplateReplacer `json:"tpContent"`
}

type TemplateReplacer struct {
	Code string `json:"valid_code"`
}

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
}

func NewClient() Client {
	return Client{
		credentials: config.MustSMSCredentials(),
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
		Signature:       "",
		TemplateID:      0,
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
	var result MessageResponse

	errs := fetch.New().
		Post("https://api.mix2.zthysms.com/v2/sendSmsTp").
		SendJSON(c.templateMessage(v)).
		EndJSON(&result)

	if errs != nil {
		return MessageResponse{}, errs[0]
	}

	if result.Valid() {
		return result, nil
	}

	return MessageResponse{}, errors.New(result.Message)
}
