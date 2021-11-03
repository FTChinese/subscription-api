package wxoauth

import "github.com/FTChinese/subscription-api/pkg/wxlogin"

// SaveWxStatus saves Wechat OAuth API error response into data so that we could know what errors on hell Wechat actually produced.
func (env Env) SaveWxStatus(rs wxlogin.RespStatus) error {

	_, err := env.dbs.Write.NamedExec(
		wxlogin.StmtSaveWxRespError,
		rs,
	)

	if err != nil {
		return err
	}

	return nil
}
