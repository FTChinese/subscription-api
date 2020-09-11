package giftrepo

import (
	"github.com/FTChinese/subscription-api/pkg/redeem"
)

func (env GiftEnv) FindGiftCard(code string) (redeem.GiftCard, error) {

	var c redeem.GiftCard
	err := env.db.Get(&c, redeem.StmtGiftCard, code)

	if err != nil {
		return c, err
	}

	return c, nil
}
