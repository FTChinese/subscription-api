package wechat

// DesktopPay is used to create response for payment in desktop browsers.
// Only the CodeURL is required by wechat.
type DesktopPay struct {
	Pay
	CodeURL string `json:"codeUrl"`
}

// MobilePay is used to create response for payment inside
// inside browsers on mobile devices.
// This url is used to redirect to a wechat web page which can
// call wechat app.
type MobilePay struct {
	Pay
	MWebURL string `json:"mWebUrl"`
}
