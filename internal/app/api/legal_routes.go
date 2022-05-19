package api

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/pkg/legal"
	"github.com/FTChinese/subscription-api/internal/repository"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/xhttp"
	"go.uber.org/zap"
	"net/http"
)

type LegalRoutes struct {
	repo   repository.LegalRepo
	logger *zap.Logger
}

func NewLegalRepo(dbs db.ReadWriteMyDBs, logger *zap.Logger) LegalRoutes {
	return LegalRoutes{
		repo:   repository.NewLegalRepo(dbs, logger),
		logger: nil,
	}
}

func (routes LegalRoutes) ListActive(w http.ResponseWriter, req *http.Request) {
	p := gorest.GetPagination(req)

	list, err := routes.repo.ListLegal(p, true)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(list)
}

func (routes LegalRoutes) ListAll(w http.ResponseWriter, req *http.Request) {
	p := gorest.GetPagination(req)

	list, err := routes.repo.ListLegal(p, false)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(list)
}

func (routes LegalRoutes) Load(w http.ResponseWriter, req *http.Request) {
	id, _ := xhttp.GetURLParam(req, "id").ToString()

	doc, err := routes.repo.Retrieve(id)

	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(doc)
}

func (routes LegalRoutes) Create(w http.ResponseWriter, req *http.Request) {
	var params legal.ContentParams
	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	legalDoc := legal.NewLegal(params)

	err := routes.repo.Create(legalDoc)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(legalDoc)
}

func (routes LegalRoutes) Update(w http.ResponseWriter, req *http.Request) {
	id, _ := xhttp.GetURLParam(req, "id").ToString()

	var params legal.ContentParams
	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	legalDoc, err := routes.repo.Retrieve(id)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	legalDoc = legalDoc.Update(params)
	err = routes.repo.Update(legalDoc)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(legalDoc)
}
