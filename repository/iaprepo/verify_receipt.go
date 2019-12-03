package iaprepo

import (
	"encoding/json"
	"gitlab.com/ftchinese/subscription-api/models/apple"
)

func (env IAPEnv) VerifyReceipt(r apple.VerificationRequestBody) (apple.VerificationResponseBody, error) {
	pw, err := getReceiptPassword()
	if err != nil {
		return apple.VerificationResponseBody{}, err
	}

	r.Password = pw
	r.ExcludeOldTransactions = false

	url := env.c.GetReceiptVerificationURL()
	logger.
		WithField("trace", "IAPEnv.VerifyReceipt").
		WithField("Using verification url %s", url)

	_, body, errs := request.
		Post(url).
		Send(r).End()

	if errs != nil {
		logger.WithField("trace", "IAPEnv.VerifyReceipt").Error(errs)
		return apple.VerificationResponseBody{}, errs[0]
	}

	logger.WithField("trace", "IAPEnv.VerifyReceipt.ResponseBody").
		Info(body)

	var resp apple.VerificationResponseBody
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		logger.WithField("trace", "IAPEnv.VerifyReceipt").Error(err)
		return resp, err
	}

	return resp, nil
}

func (env IAPEnv) SaveVerificationSession(v apple.VerificationSessionSchema) error {
	_, err := env.db.NamedExec(insertVerificationSession, v)

	if err != nil {
		logger.WithField("trace", "IAPEnv.SaveVerificationSession").Error(err)
		return err
	}

	return nil
}

func (env IAPEnv) SaveVerificationFailure(f apple.VerificationFailed) error {
	_, err := env.db.NamedExec(insertVerificationFailure, f)

	if err != nil {
		logger.WithField("trace", "IAPEnv.SaveVerificationFailure").Error(err)
		return err
	}

	return nil
}

func (env IAPEnv) SaveCustomerReceipt(r apple.TransactionSchema) error {
	_, err := env.db.NamedExec(insertTransaction, r)

	if err != nil {
		logger.WithField("trace", "IAPEnv.SaveCustomerReceipt").Error(err)
		return err
	}

	return nil
}

func (env IAPEnv) SavePendingRenewal(p apple.PendingRenewalSchema) error {
	_, err := env.db.NamedExec(insertPendingRenewal, p)

	if err != nil {
		logger.WithField("trace", "IAPEnv.SavePendingRenewal").Error(err)
		return err
	}

	return nil
}

func (env IAPEnv) CreateSubscription(s apple.Subscription) error {
	_, err := env.db.NamedExec(insertIAPSubscription, s)

	if err != nil {
		logger.WithField("trace", "IAPEnv.CreateSubscription").Error(err)
		return err
	}

	return nil
}

func (env IAPEnv) SaveReceiptToken(r apple.ReceiptToken) error {
	_, err := env.db.NamedExec(insertReceiptToken, r)

	if err != nil {
		logger.WithField("trace", "IAPEnv.SaveReceiptToken").Error(err)

		return err
	}

	return nil
}
