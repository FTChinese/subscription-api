//go:build !production
// +build !production

package test

import (
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/ali"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/postman"
	"github.com/FTChinese/subscription-api/pkg/wechat"
	"github.com/jmoiron/sqlx"
)

var (
	DB       *sqlx.DB
	SplitDB  db.ReadWriteMyDBs
	Postman  postman.Postman
	WxPayApp wechat.PayApp
	AliApp   ali.App
)

func init() {
	faker.MustSetupViper()

	SplitDB = db.MustNewMyDBs(false)
	DB = SplitDB.Write
	Postman = postman.New(config.MustGetHanqiConn())
	WxPayApp = wechat.MustNewPayApp("wxapp.app_pay")
	AliApp = ali.MustInitApp()
}
