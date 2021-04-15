package controller

import (
	"github.com/FTChinese/go-rest/postoffice"
	"github.com/FTChinese/subscription-api/pkg/db"
	"go.uber.org/zap"
)

type AuthRouter struct {
	UserShared
}

func NewAuthRouter(myDBs db.ReadWriteSplit, postman postoffice.PostOffice, l *zap.Logger) AuthRouter {
	return AuthRouter{
		UserShared: NewUserShared(myDBs, postman, l),
	}
}
