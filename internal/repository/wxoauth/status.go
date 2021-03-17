package wxoauth

import "github.com/FTChinese/subscription-api/pkg/wxlogin"

// SaveWxStatus saves Wechat OAuth API error response into data so that we could know what errors on hell Wechat actually produced.
func (env Env) SaveWxStatus(code int64, message string) error {

	_, err := env.dbs.Write.Exec(wxlogin.StmtInsertStatus,
		code,
		message,
	)

	if err != nil {
		return err
	}

	return nil
}
