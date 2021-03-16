package iaprepo

import (
	"github.com/FTChinese/subscription-api/pkg/apple"
)

func (env Env) SaveWebhook(w apple.WebHookSchema) error {
	_, err := env.dbs.Write.NamedExec(apple.StmtLoggingWebhook, w)

	if err != nil {
		return err
	}

	return nil
}
