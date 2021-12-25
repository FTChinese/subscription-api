package wxpay

import (
	"fmt"
	"go.uber.org/zap"
)

type ClientStore struct {
	clients       []Client
	indexBySource map[Source]int
	indexByID     map[string]int
}

func NewClientStore(logger *zap.Logger) ClientStore {
	store := ClientStore{
		clients:       make([]Client, 0),
		indexBySource: make(map[Source]int),
		indexByID:     make(map[string]int),
	}

	for i, cfg := range mustGetAppConfigs() {

		client := NewClient(cfg, logger)

		store.clients = append(store.clients, client)

		store.indexBySource[cfg.Platform] = i

		if cfg.Platform == SourceDesktop {
			store.indexBySource[SourceMobile] = i
		}
		store.indexByID[cfg.AppID] = i
	}

	return store
}

func (store ClientStore) SelectBySource(s Source) (Client, error) {
	i, ok := store.indexBySource[s]
	if !ok {
		return Client{}, fmt.Errorf("wxpay client: cannot find app for trade type %s", s)
	}

	return store.clients[i], nil
}

func (store ClientStore) SelectByAppID(id string) (Client, error) {
	i, ok := store.indexByID[id]

	if !ok {
		return Client{}, fmt.Errorf("wxpay client: cannot find app %s", id)
	}

	return store.clients[i], nil
}
