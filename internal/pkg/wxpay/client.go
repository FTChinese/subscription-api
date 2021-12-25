package wxpay

import (
	"github.com/go-pay/gopay/wechat"
	"go.uber.org/zap"
)

type Client struct {
	app    AppConfig
	sdk    *wechat.Client
	logger *zap.Logger
}

func NewClient(cfg AppConfig, logger *zap.Logger) Client {
	return Client{
		app:    cfg,
		sdk:    wechat.NewClient(cfg.AppID, cfg.MchID, cfg.APIKey, true),
		logger: logger,
	}
}
