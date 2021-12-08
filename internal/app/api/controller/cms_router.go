package controller

import (
	"github.com/FTChinese/subscription-api/internal/repository/cmsrepo"
	"github.com/FTChinese/subscription-api/internal/repository/shared"
	"github.com/FTChinese/subscription-api/pkg/letter"
	"go.uber.org/zap"
)

type CMSRouter struct {
	Repo         cmsrepo.Env
	ReaderRepo   shared.ReaderCommon
	PaywallRepo  shared.PaywallCommon
	Logger       *zap.Logger
	EmailService letter.Service
	Live         bool
}
