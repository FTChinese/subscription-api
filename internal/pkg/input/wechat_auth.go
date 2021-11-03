package input

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/lib/validator"
)

// WechatAccessParams contains the code acquired from
// wechat oauth workflow step 1.
type WechatAccessParams struct {
	Code string `json:"code"`
}

func (p *WechatAccessParams) Validate() *render.ValidationError {
	return validator.
		New("code").
		Required().
		Validate(p.Code)
}

// WechatRefreshParams is used the parse request body
// when user want to refresh wechat userinfo.
type WechatRefreshParams struct {
	SessionID string `json:"sessionId"`
}

func (p *WechatRefreshParams) Validate() *render.ValidationError {
	return validator.
		New("sessionId").
		Required().
		Validate(p.SessionID)
}
