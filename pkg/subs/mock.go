// +build !production

package subs

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/rand"
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

func MockProratedOrder() ProratedOrder {
	return ProratedOrder{
		OrderID:        MustGenerateOrderID(),
		Balance:        float64(rand.IntRange(10, 259)),
		CreatedUTC:     chrono.TimeNow(),
		ConsumedUTC:    chrono.Time{},
		UpgradeOrderID: "",
	}
}

func MockProratedOrderN(n int) []ProratedOrder {
	upID := MustGenerateOrderID()

	pos := make([]ProratedOrder, 0)

	for i := 0; i < n; i++ {
		o := MockProratedOrder()
		o.ConsumedUTC = chrono.TimeNow()
		o.UpgradeOrderID = upID

		pos = append(pos, o)
	}

	return pos
}
