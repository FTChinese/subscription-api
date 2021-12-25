package wechat

import (
	"fmt"
	"go.uber.org/zap"
)

// WxPayClientStore collects various wechat payment app
// in one place.
type WxPayClientStore struct {
	clients         []WxPayClient
	indexByPlatform map[TradeType]int
	indexByID       map[string]int
	logger          *zap.Logger
}

func NewWxClientStore(apps []PayApp, logger *zap.Logger) WxPayClientStore {
	store := WxPayClientStore{
		clients:         make([]WxPayClient, 0),
		indexByPlatform: make(map[TradeType]int),
		indexByID:       make(map[string]int),
		logger:          logger,
	}

	for i, app := range apps {
		store.clients = append(store.clients, NewWxPayClient(app, logger))

		store.indexByPlatform[app.Platform] = i
		// Desktop and mobile browser use the same app.
		if app.Platform == TradeTypeDesktop {
			store.indexByPlatform[TradeTypeMobile] = i
		}

		store.indexByID[app.AppID] = i
	}

	return store
}

// FindByPlatform tries to find the client used for a certain trade type.
// This is used when use is creating an order.
func (s WxPayClientStore) FindByPlatform(t TradeType) (WxPayClient, error) {
	i, ok := s.indexByPlatform[t]
	if !ok {
		return WxPayClient{}, fmt.Errorf("wxpay client: cannot find app for trade type %s", t)
	}

	return s.clients[i], nil
}

// FindByAppID searches a wechat pay app by id.
// This is used by webhook.
func (s WxPayClientStore) FindByAppID(id string) (WxPayClient, error) {
	i, ok := s.indexByID[id]

	if !ok {
		return WxPayClient{}, fmt.Errorf("wxpay client: cannot find app %s", id)
	}

	return s.clients[i], nil
}
