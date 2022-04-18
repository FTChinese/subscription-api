package api

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/xhttp"
	"net/http"
)

// InvoiceHasCoupon checks if an invoice has coupon applied.
// An invoice could only use one coupon.
func (routes StripeRoutes) InvoiceHasCoupon(w http.ResponseWriter, req *http.Request) {
	invID, _ := xhttp.GetURLParam(req, "id").ToString()

	ok, err := routes.stripeRepo.InvoiceHasCouponApplied(invID)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(map[string]bool{
		"exists": ok,
	})
}
