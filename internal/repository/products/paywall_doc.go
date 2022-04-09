package products

import (
	"github.com/FTChinese/subscription-api/pkg/reader"
)

func (env Env) CreatePaywallDoc(pwb reader.PaywallDoc) (int64, error) {
	result, err := env.dbs.Write.NamedExec(
		reader.StmtInsertPaywallDoc,
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
