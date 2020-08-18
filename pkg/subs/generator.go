package subs

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/rand"
	"strings"
)

// GenerateOrderID creates an order memberID.
// The memberID has a total length of 18 chars.
// If we use this generator:
// `FT` takes 2, followed by year-month-date-hour-minute
// FT201905191139, then only 4 chars left for random number
// 2^16 = 65536, which means only 60000 order could be created every minute.
// To leave enough space for random number, 8 chars might be reasonable - 22 chars totally.
// If we use current random number generator:
// 2 ^ 64 = 1.8 * 10^19 orders.
func GenerateOrderID() (string, error) {

	id, err := gorest.RandomHex(8)
	if err != nil {
		return "", err
	}

	return "FT" + strings.ToUpper(id), nil
}

func GenerateSnapshotID() string {
	return "snp_" + rand.String(12)
}

func GenerateUpgradeID() string {
	return "up_" + rand.String(12)
}
