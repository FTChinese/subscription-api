package query

const activateGiftCard = `
UPDATE %s.scratch_card
	SET active_time = UNIX_TIMESTAMP()
WHERE auth_code = ?
LIMIT 1`
