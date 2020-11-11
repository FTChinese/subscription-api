package poll

import (
	"github.com/FTChinese/subscription-api/pkg/poller"
	"github.com/jmoiron/sqlx"
)

type Repo struct {
	db *sqlx.DB
}

func savePollerLog(db *sqlx.DB, l *poller.Log) error {
	_, err := db.NamedExec(poller.StmtSaveLog, l)
	if err != nil {
		return err
	}

	return nil
}
