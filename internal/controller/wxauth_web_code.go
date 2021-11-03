package controller

import (
	"net/http"
	"net/url"
	"strings"
)

// WebCallback is used to help web app to get OAuth 2.0 code.
// The code and state is transferred back to next-user web app since Wechat only recognize the ftacacemy.cn URL.
func (router WxAuthRouter) WebCallback(w http.ResponseWriter, req *http.Request) {

	err := req.ParseForm()
	if err != nil {
		query := url.Values{}
		query.Set("error", "invalid_request")

		http.Redirect(
			w,
			req,
			wxOAuthCallback+query.Encode(),
			http.StatusFound)
		return
	}
	// The code returned by wechat
	code := req.FormValue("code")
	// The nonce code we send to wechat.
	state := req.FormValue("state")

	code = strings.TrimSpace(code)

	if code == "" && state != "" {
		query := url.Values{}
		query.Set("error", "access_denied")
		http.Redirect(
			w,
			req,
			wxOAuthCallback+query.Encode(),
			http.StatusFound)
		return
	}

	http.Redirect(
		w,
		req,
		wxOAuthCallback+req.Form.Encode(),
		http.StatusFound)
}
