package api

import (
	"github.com/FTChinese/subscription-api/internal/pkg/letter"
	"github.com/FTChinese/subscription-api/internal/repository/cmsrepo"
	"github.com/FTChinese/subscription-api/internal/repository/shared"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
)

type CMSRouter struct {
	repo         cmsrepo.Env
	readerRepo   shared.ReaderCommon
	emailService letter.Service
	PaymentShared
}

func NewCMSRouter(dbs db.ReadWriteMyDBs, c *cache.Cache, logger *zap.Logger, live bool) CMSRouter {
	return CMSRouter{
		repo:          cmsrepo.New(dbs, logger),
		readerRepo:    shared.NewReaderCommon(dbs),
		emailService:  letter.NewService(logger),
		PaymentShared: NewPaymentShared(dbs, c, logger, live),
	}
}
