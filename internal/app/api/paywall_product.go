package api

import (
	"net/http"

	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/xhttp"
)

func (router PaywallRouter) ListProducts(w http.ResponseWriter, req *http.Request) {
	products, err := router.productRepo.ListProducts(router.live)
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
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	var params reader.ProductParams
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

	p := reader.NewProduct(params, router.live)

	err := router.productRepo.CreateProduct(p)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(p)
}

// LoadProduct loads a single product by id.
func (router PaywallRouter) LoadProduct(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	id, _ := xhttp.GetURLParam(req, "id").ToString()

	sugar.Infof("Retrieving product %s", id)

	prod, err := router.productRepo.RetrieveProduct(id, router.live)
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
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	id, _ := xhttp.GetURLParam(req, "id").ToString()

	var params reader.ProductParams
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

	sugar.Infof("Retrieving product %s", id)

	prod, err := router.productRepo.RetrieveProduct(id, router.live)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	sugar.Infof("Product retrieved %v", prod)

	updated := prod.Update(params)
	err = router.productRepo.UpdateProduct(updated)
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

	prod, err := router.productRepo.RetrieveProduct(id, router.live)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	prod = prod.Activate()
	err = router.productRepo.SetProductOnPaywall(prod)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(prod)
}

// AttachIntroPrice activate an existing one_time price
// and attaches it to the product it belongs to if not attached yet,
// or override existing introductory price of a product.
// Request body:
// - priceId: string;
func (router PaywallRouter) AttachIntroPrice(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	id, err := xhttp.GetURLParam(req, "id").ToString()
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		sugar.Error(err)
		return
	}

	var params reader.ProductIntroParams
	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		sugar.Error(err)
		return
	}

	pwPrice, err := router.paywallRepo.RetrievePaywallPrice(
		params.PriceID,
		router.live)

	if err != nil {
		_ = render.New(w).DBError(err)
		sugar.Error(err)
		return
	}

	// If this price is not of type one_time
	if !pwPrice.IsOneTime() {
		_ = render.New(w).Unprocessable(&render.ValidationError{
			Message: "Only one_time price could be used for introductory",
			Field:   "priceId",
			Code:    render.CodeInvalid,
		})
		return
	}

	activated := pwPrice.Activate()
	// If the price is not activated yet.
	if !pwPrice.Active {
		err = router.productRepo.ActivatePrice(activated)
		if err != nil {
			_ = render.New(w).DBError(err)
			sugar.Error(err)
			return
		}
	}

	prod, err := router.productRepo.RetrieveProduct(id, router.live)
	if err != nil {
		_ = render.New(w).DBError(err)
		sugar.Error(err)
		return
	}

	prod = prod.WithIntroPrice(activated)

	err = router.productRepo.SetProductIntro(prod)
	if err != nil {
		_ = render.New(w).DBError(err)
		sugar.Error(err)
		return
	}

	_ = render.New(w).OK(prod)
}

func (router PaywallRouter) DropIntroPrice(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	id, err := xhttp.GetURLParam(req, "id").ToString()
	if err != nil {
		_ = render.New(w).BadRequest(err.Error())
		sugar.Error(err)
		return
	}

	prod, err := router.productRepo.RetrieveProduct(id, router.live)
	if err != nil {
		_ = render.New(w).DBError(err)
		sugar.Error(err)
		return
	}

	err = router.productRepo.DeactivatePrice(prod.Introductory.Deactivate(false))
	if err != nil {
		_ = render.New(w).DBError(err)
		sugar.Error(err)
		return
	}

	prod = prod.DropIntroPrice()
	err = router.productRepo.SetProductIntro(prod)
	if err != nil {
		_ = render.New(w).DBError(err)
		sugar.Error(err)
		return
	}

	_ = render.New(w).OK(prod)
}
