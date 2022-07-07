package api

import (
	"database/sql"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/pkg/stripe"
	"github.com/FTChinese/subscription-api/pkg/xhttp"
	"net/http"
)

func (routes StripeRoutes) loadInvoice(id string, refresh bool) (stripe.Invoice, error) {
	defer routes.logger.Sync()
	sugar := routes.logger.Sugar()

	inv, err := routes.stripeRepo.LoadOrFetchInvoice(id, refresh)
	if err != nil {
		return stripe.Invoice{}, err
	}

	if inv.IsFromStripe {
		go func() {
			err := routes.stripeRepo.UpsertInvoice(inv)
			if err != nil {
				sugar.Error(err)
			}
		}()
	}

	return inv, nil
}

func (routes StripeRoutes) LoadLatestInvoice(w http.ResponseWriter, req *http.Request) {
	defer routes.logger.Sync()
	sugar := routes.logger.Sugar()

	subsID, err := xhttp.GetURLParam(req, "id").ToString()
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	subs, err := routes.stripeRepo.LoadOrFetchSubs(subsID, false)
	if err != nil {
		sugar.Error(err)
		_ = xhttp.HandleSubsErr(w, err)
		return
	}

	inv, err := routes.loadInvoice(subs.LatestInvoiceID, false)
	if err != nil {
		sugar.Error(err)
		_ = xhttp.HandleSubsErr(w, err)
		return
	}

	_ = render.New(w).OK(inv)
}

func (routes StripeRoutes) LoadInvoice(w http.ResponseWriter, req *http.Request) {
	defer routes.logger.Sync()
	sugar := routes.logger.Sugar()

	refresh := xhttp.ParseQueryRefresh(req)

	id, err := xhttp.GetURLParam(req, "id").ToString()
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	inv, err := routes.loadInvoice(id, refresh)
	if err != nil {
		sugar.Error(err)
		_ = xhttp.HandleSubsErr(w, err)
		return
	}

	_ = render.New(w).OK(inv)
}

func (routes StripeRoutes) CouponOfLatestInvoice(w http.ResponseWriter, req *http.Request) {
	defer routes.logger.Sync()
	sugar := routes.logger.Sugar()

	subsID, err := xhttp.GetURLParam(req, "id").ToString()
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	subs, err := routes.stripeRepo.LoadOrFetchSubs(subsID, false)
	if err != nil {
		sugar.Error(err)
		_ = xhttp.HandleSubsErr(w, err)
		return
	}

	applied, err := routes.stripeRepo.LatestCouponApplied(subs.LatestInvoiceID)
	if err != nil {
		if err == sql.ErrNoRows {
			_ = render.New(w).OK(stripe.CouponRedeemed{})
			return
		}

		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(applied)
}
