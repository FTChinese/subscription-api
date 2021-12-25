package controller

import (
	"github.com/FTChinese/subscription-api/internal/pkg/letter"
	"github.com/FTChinese/subscription-api/internal/repository/cmsrepo"
	"github.com/FTChinese/subscription-api/internal/repository/shared"
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
