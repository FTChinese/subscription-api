package api

import (
	"github.com/FTChinese/subscription-api/internal/pkg/letter"
	"github.com/FTChinese/subscription-api/internal/repository"
	"github.com/FTChinese/subscription-api/internal/repository/cmsrepo"
	"github.com/FTChinese/subscription-api/internal/repository/shared"
	"github.com/FTChinese/subscription-api/pkg/db"
	"go.uber.org/zap"
)

type CMSRouter struct {
	repo         cmsrepo.Env
	readerRepo   shared.ReaderCommon
	paywallRepo  repository.PaywallRepo
	emailService letter.Service
	logger       *zap.Logger
	live         bool
}

func NewCMSRouter(dbs db.ReadWriteMyDBs, live bool, logger *zap.Logger) CMSRouter {
	return CMSRouter{
		repo:         cmsrepo.New(dbs, logger),
		readerRepo:   shared.NewReaderCommon(dbs),
		paywallRepo:  repository.NewPaywallRepo(dbs),
		emailService: letter.NewService(logger),
		live:         live,
		logger:       logger,
	}
}
