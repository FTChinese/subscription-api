package giftrepo

import (
	"gitlab.com/ftchinese/subscription-api/models/redeem"
	"gitlab.com/ftchinese/subscription-api/repository/query"
)

func (env GiftEnv) FindGiftCard(code string) (redeem.GiftCard, error) {

	var c redeem.GiftCard
	err := env.db.Get(&c, query.RetrieveGiftCard, code)

	if err != nil {
		logger.WithField("trace", "FindGiftCard").Error(err)
		return c, err
	}

	return c, nil
}
