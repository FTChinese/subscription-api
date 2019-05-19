package query

import "fmt"

func (b Builder) ActivateGiftCard() string {
	return fmt.Sprintf(`
	UPDATE %s.scratch_card
		SET active_time = UNIX_TIMESTAMP()
	WHERE auth_code = ?
	LIMIT 1`, b.MemberDB())
}
