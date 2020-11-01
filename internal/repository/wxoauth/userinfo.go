package wxoauth

import (
	"github.com/FTChinese/subscription-api/pkg/wxlogin"
)

// SaveWxUser from wechat API.
// Since a user can authorize multiple times, use ON DUPLICATE to handle unique key constraint.
func (env Env) SaveWxUser(u wxlogin.UserInfoSchema) error {

	_, err := env.db.NamedExec(wxlogin.StmtInsertUserInfo, u)

	if err != nil {
		return err
	}

	return nil
}

// UpdateWxUser update data of one union id.
func (env Env) UpdateWxUser(u wxlogin.UserInfo) error {
	_, err := env.db.Exec(wxlogin.StmtUpdateUserInfo,
		u.SQLSchema())

	if err != nil {
		return err
	}

	return nil
}
