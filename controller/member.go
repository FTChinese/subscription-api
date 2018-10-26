package controller

import (
	"database/sql"
	"net/http"

	"gitlab.com/ftchinese/subscription-api/model"
	"gitlab.com/ftchinese/subscription-api/util"
)

// MemberRouter checks membership status
type MemberRouter struct {
	model model.Env
}

// NewMemberRouter creates a new istance of MemberRouter
func NewMemberRouter(db *sql.DB) MemberRouter {
	return MemberRouter{
		model: model.Env{DB: db},
	}
}

// IsRenewable answers if user is allowed to renew membership.
func (mr MemberRouter) IsRenewable(w http.ResponseWriter, req *http.Request) {
	util.Render(w, util.NewNoContent())
}
