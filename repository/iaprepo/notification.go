package iaprepo

import "github.com/FTChinese/subscription-api/pkg/apple"

func (env Env) SaveNotification(w apple.WebHookSchema) error {
	_, err := env.db.NamedExec(apple.StmtLoggingWebhook, w)

	if err != nil {
		logger.WithField("trace", "Env.SaveNotification").Error(err)

		return err
	}

	return nil
}
