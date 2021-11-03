package wxoauth

import (
	"github.com/FTChinese/subscription-api/pkg/wxlogin"
)

// UpsertUserInfo from wechat API.
// Since a user can authorize multiple times,
// use ON DUPLICATE to handle unique key constraint.
func (env Env) UpsertUserInfo(u wxlogin.UserInfoSchema) error {

	_, err := env.dbs.Write.NamedExec(
		wxlogin.StmtUpsertUserInfo,
		u)

	if err != nil {
		return err
	}

	return nil
}

// UpdateWxUser update data of one union id.
// Deprecated
func (env Env) UpdateWxUser(u wxlogin.UserInfoSchema) error {
	_, err := env.dbs.Write.Exec(
		wxlogin.StmtUpdateUserInfo,
		u)

	if err != nil {
		return err
	}

	return nil
}
