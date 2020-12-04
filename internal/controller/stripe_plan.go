package controller

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/product"
	"net/http"
	"strings"
)

// GetPlan retrieves a stripe plan by id.
// GET /stripe/plans/<standard_month | standard_year | premium_year>
// Deprecated
func (router StripeRouter) GetPlan(w http.ResponseWriter, req *http.Request) {
	namedKey, err := getURLParam(req, "id").ToString()
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	parts := strings.Split(namedKey, "_")
	if len(parts) < 2 {
		_ = render.New(w).NotFound()
		return
	}

	tier, err := enum.ParseTier(parts[0])
	cycle, err := enum.ParseCycle(parts[1])
	if tier == enum.TierNull || cycle == enum.CycleNull {
		_ = render.New(w).NotFound()
		return
	}

	// Fetch plan from Stripe API
	p, err := router.client.GetPlan(product.Edition{
		Tier:  tier,
		Cycle: cycle,
	})

	if err != nil {
		err = forwardStripeErr(w, err)
		if err == nil {
			return
		}

		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(p)
}

func (router StripeRouter) GetPrice(w http.ResponseWriter, req *http.Request) {
	edition, err := getEdition(req)
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	sp, err := router.client.GetPlan(edition)
	if err != nil {
		err = forwardStripeErr(w, err)
		if err == nil {
			return
		}

		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(sp)
}
