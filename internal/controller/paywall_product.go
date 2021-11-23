package controller

import (
	"github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/pw"
	"net/http"
)

func (router PaywallRouter) ListProducts(w http.ResponseWriter, req *http.Request) {
	products, err := router.repo.ListProducts(router.live)
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
	var params pw.ProductParams
	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	if ve := params.Validate(false); ve != nil {
		_ = render.New(w).Unprocessable(ve)
		return
	}

	p := pw.NewProduct(params, router.live)

	err := router.repo.CreateProduct(p)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(p)
}

// LoadProduct loads a single product by id.
func (router PaywallRouter) LoadProduct(w http.ResponseWriter, req *http.Request) {
	id, _ := getURLParam(req, "id").ToString()

	prod, err := router.repo.RetrieveFtcPrice(id, router.live)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(prod)
}

func (router PaywallRouter) UpdateProduct(w http.ResponseWriter, req *http.Request) {
	id, _ := getURLParam(req, "id").ToString()

	var params pw.ProductParams
	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	if ve := params.Validate(true); ve != nil {
		_ = render.New(w).Unprocessable(ve)
		return
	}

	prod, err := router.repo.RetrieveProduct(id, router.live)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	updated := prod.Update(params)
	err = router.repo.UpdateProduct(updated)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(updated)
}

func (router PaywallRouter) ActivateProduct(w http.ResponseWriter, req *http.Request) {
	id, err := getURLParam(req, "id").ToString()
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	prod, err := router.repo.RetrieveProduct(id, router.live)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	prod = prod.Activate()
	err = router.repo.SetProductOnPaywall(prod)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(prod)
}
