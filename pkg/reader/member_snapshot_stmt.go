package reader

const StmtSnapshotMember = `
INSERT INTO premium.member_snapshot
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
