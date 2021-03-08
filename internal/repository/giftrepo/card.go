package giftrepo

import (
	"github.com/FTChinese/subscription-api/pkg/giftcard"
)

func (env GiftEnv) FindGiftCard(code string) (giftcard.GiftCard, error) {

	var c giftcard.GiftCard
	err := env.db.Get(&c, giftcard.StmtGiftCard, code)

	if err != nil {
		return c, err
	}

	return c, nil
}
