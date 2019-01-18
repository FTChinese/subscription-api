package controller

import (
	"net/http"

	"github.com/FTChinese/go-rest/view"
	"gitlab.com/ftchinese/subscription-api/model"
)

// MemberRouter checks membership status
type MemberRouter struct {
	model model.Env
}

// NewMemberRouter creates a new istance of MemberRouter
func NewMemberRouter(m model.Env) MemberRouter {
	return MemberRouter{
		model: m,
	}
}

// IsRenewable answers if user is allowed to renew membership.
func (mr MemberRouter) IsRenewable(w http.ResponseWriter, req *http.Request) {
	view.Render(w, view.NewNoContent())
}
