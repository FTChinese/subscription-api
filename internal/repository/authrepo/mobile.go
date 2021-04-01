package authrepo

import (
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/ztsms"
)

func (env Env) SaveSMSVerifier(v ztsms.Verifier) error {
	_, err := env.dbs.Write.NamedExec(ztsms.StmtSaveVerifier, v)
	if err != nil {
		return err
	}

	return nil
}

func (env Env) RetrieveSMSVerifier(params ztsms.VerifierParams) (ztsms.Verifier, error) {
	var v ztsms.Verifier
	err := env.dbs.Read.Get(&v, ztsms.StmtRetrieveVerifier, params.Mobile, params.Code)
	if err != nil {
		return ztsms.Verifier{}, err
	}

	return v, nil
}

func (env Env) SMSVerifierUsed(v ztsms.Verifier) error {
	_, err := env.dbs.Write.NamedExec(ztsms.StmtVerifierUsed, v)

	if err != nil {
		return err
	}

	return nil
}

func (env Env) UserIDByPhone(phone string) (string, error) {
	var id string
	err := env.dbs.Read.Get(&id, ztsms.StmtUserIDByPhone, phone)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (env Env) SetPhone(a account.Ftc) error {
	_, err := env.dbs.Write.NamedExec(ztsms.StmtSetPhone, a)
	if err != nil {
		return err
	}

	return nil
}
