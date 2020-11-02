// +build !production

package subs

import (
	"github.com/FTChinese/go-rest/chrono"
	"time"
)

func MockBalanceSource() BalanceSource {
	return BalanceSource{
		OrderID:   MustGenerateOrderID(),
		Amount:    258.00,
		StartDate: chrono.DateFrom(time.Now()),
		EndDate:   chrono.DateFrom(time.Now().AddDate(1, 0, 1)),
	}
}

func MockBalanceSourceN(n int) []BalanceSource {
	bs := make([]BalanceSource, 0)
	for i := 0; i < n; i++ {
		bs = append(bs, MockBalanceSource())
	}

	return bs
}
