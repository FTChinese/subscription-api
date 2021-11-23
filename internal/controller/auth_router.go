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
	l *zap.Logger,
	pm postman.Postman,
) AuthRouter {
	return AuthRouter{
		UserShared: NewUserShared(myDBs, pm, l),
	}
}
