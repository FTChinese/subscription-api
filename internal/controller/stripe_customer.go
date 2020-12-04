package controller

import (
	"github.com/FTChinese/go-rest/render"
	"net/http"
)

// CreateCustomer creates stripe customer if not present.
// PUT /stripe/customers
// Response: reader.FtcAccount
func (router StripeRouter) CreateCustomer(w http.ResponseWriter, req *http.Request) {
	ftcID := req.Header.Get(userIDKey)

	account, err := router.stripeRepo.CreateCustomer(ftcID)

	if err != nil {
		err := forwardStripeErr(w, err)
		if err != nil {
			_ = render.New(w).DBError(err)
		}

		return
	}

	_ = render.New(w).OK(account)
}
