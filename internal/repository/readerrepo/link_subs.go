package readerrepo

import "github.com/FTChinese/subscription-api/pkg/reader"

func (env Env) LinkSubs(l reader.SubsLink) error {
	_, err := env.db.NamedExec(reader.StmtSaveSubsLink, l)
	if err != nil {
		return err
	}

	return nil
}
