package redeem

import (
	"fmt"
	"github.com/FTChinese/subscription-api/pkg/config"
)

const StmtGiftCard = `
SELECT auth_code AS redeemCode,
	tier AS tier,
	cycle_unit AS cycleUnit,
	cycle_value AS cycleValue
FROM premium.scratch_card
WHERE auth_code = ?
	AND expire_time > UNIX_TIMESTAMP()
	AND active_time = 0
LIMIT 1`

const activateGiftCard = `
UPDATE %s.scratch_card
	SET active_time = UNIX_TIMESTAMP()
WHERE auth_code = ?
LIMIT 1`

func StmtActivateGiftCard(db config.SubsDB) string {
	return fmt.Sprintf(activateGiftCard, db)
}
