package test

import (
	"fmt"
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/enum"
	"github.com/Pallinder/go-randomdata"
	"github.com/guregu/null"
	"gitlab.com/ftchinese/subscription-api/paywall"
	"strings"
	"time"
)

func GenCardSerial() string {
	now := time.Now()
	anni := now.Year() - 2005
	suffix := randomdata.Number(0, 9999)

	return fmt.Sprintf("%d%02d%04d", anni, now.Month(), suffix)
}

func MockGiftCard() paywall.GiftCard {
	code, _ := gorest.RandomHex(8)

	return paywall.GiftCard{
		Code:       strings.ToUpper(code),
		Tier:       enum.TierStandard,
		CycleUnit:  enum.CycleYear,
		CycleValue: null.IntFrom(1),
	}
}

func CreateGiftCard() paywall.GiftCard {
	c := MockGiftCard()

	query := `
	INSERT INTO premium.scratch_card
		SET serial_number = ?,
			auth_code = ?,
		    expire_time = UNIX_TIMESTAMP(?),
			tier = ?,
			cycle_unit = ?,
			cycle_value = ?`

	now := time.Now()

	_, err := DB.Exec(query,
		GenCardSerial(),
		c.Code,
		now.Truncate(24*time.Hour),
		c.Tier,
		c.CycleUnit,
		c.CycleValue)

	if err != nil {
		panic(err)
	}

	return c
}
