package controller

import (
	"github.com/FTChinese/go-rest/postoffice"
	"github.com/FTChinese/subscription-api/internal/repository/accounts"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/ztsms"
	"go.uber.org/zap"
)

type AuthRouter struct {
	repo    accounts.Env
	client  ztsms.Client
	postman postoffice.PostOffice
	logger  *zap.Logger
}

func NewAuthRouter(myDBs db.ReadWriteSplit, postman postoffice.PostOffice, l *zap.Logger) AuthRouter {
	return AuthRouter{
		repo:    accounts.New(myDBs, l),
		client:  ztsms.NewClient(),
		postman: postman,
		logger:  l,
	}
}
