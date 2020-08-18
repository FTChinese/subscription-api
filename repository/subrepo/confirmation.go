package subrepo

import (
	"github.com/FTChinese/subscription-api/pkg/subs"
)

func (env SubEnv) SaveConfirmationResult(r subs.ConfirmErrSchema) error {
	_, err := env.db.NamedExec(
		subs.StmtSaveConfirmResult,
		r)

	if err != nil {
		return err
	}

	return nil
}
