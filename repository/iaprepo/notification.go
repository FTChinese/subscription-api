package iaprepo

import "github.com/FTChinese/subscription-api/pkg/apple"

func (env IAPEnv) SaveNotification(w apple.WebHookSchema) error {
	_, err := env.db.NamedExec(insertWebHook, w)

	if err != nil {
		logger.WithField("trace", "IAPEnv.SaveNotification").Error(err)

		return err
	}

	return nil
}
