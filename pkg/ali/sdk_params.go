package ali

import (
	"github.com/guregu/null"
)

// SDKParams contains the parameters to call alipay sdk.
// There are two approaches to call alipay depending on
// where you are using it:
// - Inside any browser, redirect to alipay;
// - Inside native app, use the parameters to call alipay sdk.
type SDKParams struct {
	BrowserRedirect null.String `json:"browserRedirect"`
	AppSDK          null.String `json:"appSdk"`
}

func (p SDKParams) IsZero() bool {
	return p.BrowserRedirect.IsZero() && p.AppSDK.IsZero()
}
