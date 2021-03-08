package giftcard

import (
	"errors"
	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
	"time"
)

// GiftCard contains information for a gift card.
type GiftCard struct {
	Code       string
	Tier       enum.Tier
	CycleUnit  enum.Cycle
	CycleValue null.Int
}

func (c GiftCard) ExpireTime() (time.Time, error) {

	today := time.Now()

	if c.CycleValue.IsZero() {
		return today, errors.New("cycle value is zero")
	}

	cycleVal := int(c.CycleValue.Int64)

	switch c.CycleUnit {
	case enum.CycleMonth:
		return today.AddDate(0, cycleVal, 1), nil

	case enum.CycleYear:
		return today.AddDate(cycleVal, 0, 1), nil

	default:
		return today, errors.New("invalid cycle")
	}
}
