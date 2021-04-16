package accounts

import (
	"github.com/FTChinese/subscription-api/pkg/footprint"
)

func (env Env) SaveFootprint(f footprint.Footprint) error {
	_, err := env.DBs.Write.NamedExec(footprint.StmtInsertFootprint, f)

	if err != nil {
		return err
	}

	return nil
}
