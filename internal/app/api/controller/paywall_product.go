package controller

import (
	"github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/pw"
	"github.com/FTChinese/subscription-api/pkg/xhttp"
	"net/http"
)

func (router PaywallRouter) ListProducts(w http.ResponseWriter, req *http.Request) {
	products, err := router.ProductRepo.ListProducts(router.Live)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(products)
}

// CreateProduct creates a new product.
// Request body:
// - createdBy: string;
// - description?: string;
// - heading: string;
// - smallPrint?: string;
// - tier: standard | premium;
func (router PaywallRouter) CreateProduct(w http.ResponseWriter, req *http.Request) {
	defer router.Logger.Sync()
	sugar := router.Logger.Sugar()

	var params pw.ProductParams
	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		sugar.Error(err)
		return
	}

	if ve := params.Validate(false); ve != nil {
		sugar.Error(ve)
		_ = render.New(w).Unprocessable(ve)
		return
	}

	p := pw.NewProduct(params, router.Live)

	if p.Introductory.StripePriceID.Valid {
		_, err := router.StripePriceRepo.LoadPrice(p.Introductory.StripePriceID.String, router.Live)
		if err != nil {
			sugar.Error(err)
			_ = render.New(w).Unprocessable(&render.ValidationError{
				Message: err.Error(),
				Field:   "introductory.stripePriceId",
				Code:    render.CodeInvalid,
			})
			return
		}
	}

	err := router.ProductRepo.CreateProduct(p)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(p)
}

// LoadProduct loads a single product by id.
func (router PaywallRouter) LoadProduct(w http.ResponseWriter, req *http.Request) {
	defer router.Logger.Sync()
	sugar := router.Logger.Sugar()

	id, _ := xhttp.GetURLParam(req, "id").ToString()

	sugar.Infof("Retrieving product %s", id)

	prod, err := router.ProductRepo.RetrieveProduct(id, router.Live)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(prod)
}

// UpdateProduct changes product content.
// Request boyd:
// - description: string;
// - heading: string;
// - smallPrint: string;
func (router PaywallRouter) UpdateProduct(w http.ResponseWriter, req *http.Request) {
	defer router.Logger.Sync()
	sugar := router.Logger.Sugar()

	id, _ := xhttp.GetURLParam(req, "id").ToString()

	var params pw.ProductParams
	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	if ve := params.Validate(true); ve != nil {
		sugar.Error(ve)
		_ = render.New(w).Unprocessable(ve)
		return
	}

	if params.Introductory.StripePriceID.Valid {
		_, err := router.StripePriceRepo.LoadPrice(params.Introductory.StripePriceID.String, router.Live)
		if err != nil {
			sugar.Error(err)
			_ = render.New(w).Unprocessable(&render.ValidationError{
				Message: err.Error(),
				Field:   "introductory.stripePriceId",
				Code:    render.CodeInvalid,
			})
			return
		}
	}

	sugar.Infof("Retrieving product %s", id)

	prod, err := router.ProductRepo.RetrieveProduct(id, router.Live)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	sugar.Infof("Product retrieved %v", prod)

	updated := prod.Update(params)
	err = router.ProductRepo.UpdateProduct(updated)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(updated)
}

func (router PaywallRouter) ActivateProduct(w http.ResponseWriter, req *http.Request) {
	id, err := xhttp.GetURLParam(req, "id").ToString()
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	prod, err := router.ProductRepo.RetrieveProduct(id, router.Live)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	prod = prod.Activate()
	err = router.ProductRepo.SetProductOnPaywall(prod)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(prod)
}
