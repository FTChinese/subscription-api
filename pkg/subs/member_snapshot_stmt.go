package subs

import (
	"fmt"
	"github.com/FTChinese/subscription-api/pkg/config"
)

const insertMemberSnapshot = `
INSERT INTO %s.member_snapshot
SET id = :snapshot_id,
	reason = :reason,
	created_utc = UTC_TIMESTAMP(),
	order_id = :order_id,
	compound_id = :compound_id,
	ftc_user_id = :ftc_id,
	wx_union_id = :union_id,
	tier = :tier,
	cycle = :cycle,
` + mUpsertSharedCols

func StmtSnapshotMember(db config.SubsDB) string {
	return fmt.Sprintf(insertMemberSnapshot, db)
}
