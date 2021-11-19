package controller

import (
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/postman"
	"go.uber.org/zap"
)

type AuthRouter struct {
	UserShared
}

func NewAuthRouter(
	myDBs db.ReadWriteMyDBs,
	pm postman.Postman,
	l *zap.Logger) AuthRouter {
	return AuthRouter{
		UserShared: NewUserShared(myDBs, pm, l),
	}
}
