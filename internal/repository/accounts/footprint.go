package accounts

import (
	"github.com/FTChinese/subscription-api/pkg/footprint"
)

func (env Env) SaveFootprint(f footprint.Footprint) error {
	_, err := env.dbs.Write.NamedExec(footprint.StmtInsertFootprint, f)

	if err != nil {
		return err
	}

	return nil
}
