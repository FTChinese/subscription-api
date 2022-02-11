package api

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/wxlogin"
	"net/http"
)

// WxCallbackHandler is used to help web app to get OAuth 2.0 code.
// The code and state is transferred back to the web app initialized the
// OAuth workflow since Wechat only recognize the www.ftacacemy.cn URL.
// * If granted, query parameter will be:
// `?code=<string>&state=<string>`
func WxCallbackHandler(app wxlogin.CallbackApp) http.HandlerFunc {

	return func(w http.ResponseWriter, req *http.Request) {
		redirectTo, err := wxlogin.GetCallbackURL(app, req.Form)
		if err != nil {
			_ = render.New(w).BadRequest(err.Error())
			return
		}

		http.Redirect(
			w,
			req,
			redirectTo,
			http.StatusFound)
	}
}
