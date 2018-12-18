package wxlogin

// BindAccount binds a wechat account to ftc account.
func (env Env) BindAccount(userID, unionID string) error {
	query := `UPDATE cmstmp01.userinfo
	SET wx_union_id = ?
	WHERE user_id = ?
	LIMIT 1`

	_, err := env.DB.Exec(query, unionID, userID)

	if err != nil {
		logger.WithField("trace", "BindAccount").Error(err)
		return err
	}

	return nil
}

// BindAccountAndMember associate a wechat account with an FTC account.
// The FTC account must not be bound to a wechat account,
// And must not subscribed to any kind of membership.
// It set the wx_union_id column to wechat unioin id and set the membership's vip_id column to user id.
func (env Env) BindAccountAndMember(merged Membership) error {
	tx, err := env.DB.Begin()
	if err != nil {
		logger.WithField("trace", "BindAccount").Error(err)
		return err
	}

	// Update the wx_union_id field of a user's account based on user id.
	stmtAccount := `
	UPDATE cmstmp01.userinfo
	SET wx_union_id = ?
	WHERE user_id = ?
	LIMIT 1`

	_, errA := tx.Exec(stmtAccount, merged.UnionID, merged.UserID)

	// Error 1062: Duplicate entry 'ogfvwjk6bFqv2yQpOrac0J3PqA0o' for key 'wx_union_id'
	if errA != nil {
		_ = tx.Rollback()
		logger.WithField("trace", "BindAccount set union id").Error(errA)
	}

	// Delete wechat membership
	stmtDelete := `
	DELETE premium.ftc_vip
	WHERE vip_id_alias = ?
	LIMIT 1`

	_, errB := tx.Exec(stmtDelete, merged.UnionID)

	if errB != nil {
		_ = tx.Rollback()
		logger.WithField("trace", "BindAccountAndMember delete wechat member").Error(errB)
	}

	stmtUpdate := `
	INSERT INTO premium.ftc_vip
	SET vip_id = ?,
		vip_id_alias = ?,
		member_tier = ?,
		billing_cycle = ?,
		expire_date = '2019-01-01'
	ON DUPLICATE KEY UPDATE
		vip_id_alias = ?,
		member_tier = ?,
		billing_cycle = ?,
		expire_date = ?`

	_, errC := tx.Exec(stmtUpdate,
		merged.UserID,
		merged.UnionID,
		merged.Tier,
		merged.Cycle,
		merged.ExpireDate,
		merged.UnionID,
		merged.Tier,
		merged.Cycle,
		merged.ExpireDate,
	)

	// Error 1062: Duplicate entry 'e1a1f5c0-0e23-11e8-aa75-977ba2bcc6ae' for key 'PRIMARY'"
	// If the `userID` is already a member.
	if errC != nil {
		_ = tx.Rollback()
		logger.WithField("trace", "BindAccount").Error(errC)
	}

	if err := tx.Commit(); err != nil {
		logger.WithField("trace", "BindAccount commit trasaction").Error(err)

		return err
	}

	return nil
}
