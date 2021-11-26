package products

import (
	"database/sql"
	"github.com/FTChinese/subscription-api/pkg/pw"
)

func (env Env) CreatePaywallDoc(pwb pw.PaywallDoc) (int64, error) {
	result, err := env.dbs.Write.NamedExec(
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

func (env Env) RetrievePaywallDoc(live bool) (pw.PaywallDoc, error) {
	var pwb pw.PaywallDoc

	err := env.dbs.Read.Get(
		&pwb,
		pw.StmtRetrievePaywallDoc,
		live)

	if err != nil {
		if err != sql.ErrNoRows {
			return pw.PaywallDoc{}, err
		}

		// No paywall doc exists yet. Returns an empty version.
		return pw.PaywallDoc{
			LiveMode: live,
		}, nil
	}

	return pwb, nil
}
