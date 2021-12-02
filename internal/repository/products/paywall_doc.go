package products

import (
	"github.com/FTChinese/subscription-api/pkg/pw"
)

func (env Env) CreatePaywallDoc(pwb pw.PaywallDoc) (int64, error) {
	result, err := env.DBs.Write.NamedExec(
		pw.StmtInsertPaywallDoc,
		pwb)

	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}
