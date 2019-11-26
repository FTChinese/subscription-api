package controller

import (
	"github.com/jmoiron/sqlx"
	"net/http"
)

type IAPRouter struct {
	password string
	db       *sqlx.DB
}

func NewIAPRouter(pw string, db *sqlx.DB) IAPRouter {
	return IAPRouter{
		password: pw,
		db:       db,
	}
}

// VerifyReceipt perform app store receipt verification
func (router IAPRouter) VerifyReceipt(w http.ResponseWriter, req *http.Request) {

}

// WebHook receives app store server-to-server notification.
func (router IAPRouter) WebHook(w http.ResponseWriter, req *http.Request) {

}
