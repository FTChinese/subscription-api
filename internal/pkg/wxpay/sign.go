package wxpay

import (
	"errors"
	"github.com/go-pay/gopay/wechat"
	"github.com/guregu/null"
)

func (app AppConfig) SignAppParams(resp *wechat.UnifiedOrderResponse) NativeAppParams {
	p := NewNativeAppParams(resp)

	p.Signature = wechat.GetAppPaySign(
		p.AppID,
		p.PartnerID,
		p.Nonce,
		p.PrepayID,
		app.SignType,
		p.Timestamp,
		app.APIKey,
	)

	return p
}

func (app AppConfig) SignJSApiParams(resp *wechat.UnifiedOrderResponse) JSApiParams {
	p := NewJSApiParams(resp, app.SignType)

	p.Signature = wechat.GetJsapiPaySign(
		p.AppID,
		p.Nonce,
		p.Package,
		app.SignType,
		p.Timestamp,
		app.APIKey,
	)

	return p
}

func (app AppConfig) NewPayParams(resp *wechat.UnifiedOrderResponse, source Source) (PayParams, error) {
	switch source {
	case SourceDesktop:
		return PayParams{
			DesktopQr:      null.StringFrom(resp.CodeUrl),
			MobileRedirect: null.String{},
			JsApi:          JSApiParamsJSON{},
			AppSDK:         NativeAppParamsJSON{},
		}, nil

	case SourceMobile:
		return PayParams{
			DesktopQr:      null.String{},
			MobileRedirect: null.StringFrom(resp.MwebUrl),
			JsApi:          JSApiParamsJSON{},
			AppSDK:         NativeAppParamsJSON{},
		}, nil

	case SourceJSAPI:
		return PayParams{
			DesktopQr:      null.String{},
			MobileRedirect: null.String{},
			JsApi: JSApiParamsJSON{
				JSApiParams: app.SignJSApiParams(resp),
			},
			AppSDK: NativeAppParamsJSON{},
		}, nil

	case SourceApp:
		return PayParams{
			DesktopQr:      null.String{},
			MobileRedirect: null.String{},
			JsApi:          JSApiParamsJSON{},
			AppSDK: NativeAppParamsJSON{
				NativeAppParams: app.SignAppParams(resp),
			},
		}, nil

	default:
		return PayParams{}, errors.New("unknown wxpay platform")
	}
}
