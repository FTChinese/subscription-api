package controller

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/repository/accounts"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/ztsms"
	"go.uber.org/zap"
	"net/http"
)

type AccountRouter struct {
	repo      accounts.Env
	smsClient ztsms.Client
	logger    *zap.Logger
}

func NewAccountRouter(myDBs db.ReadWriteSplit, l *zap.Logger) AccountRouter {
	return AccountRouter{
		repo:      accounts.New(myDBs, l),
		smsClient: ztsms.NewClient(),
		logger:    l,
	}
}

func (router AccountRouter) LoadAccountByEmail(w http.ResponseWriter, req *http.Request) {
	userID := req.Header.Get(userIDKey)

	acnt, err := router.repo.AccountByFtcID(userID)

	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(acnt)
}

// LoadAccountByWx respond to request for user account by X-Union-Id.
//
//	GET /wx/account
// Header `X-Union-Id: <wechat union id>`
func (router AccountRouter) LoadAccountByWx(w http.ResponseWriter, req *http.Request) {
	unionID := req.Header.Get(unionIDKey)

	acnt, err := router.repo.AccountByWxID(unionID)

	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(acnt)
}
