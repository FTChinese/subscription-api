//go:build !production

package legal

import (
	"github.com/FTChinese/subscription-api/faker"
	"github.com/brianvoe/gofakeit/v5"
	"github.com/guregu/null"
)

func NewMockLegal() Legal {
	faker.SeedGoFake()

	return NewLegal(
		ContentParams{
			Author:  gofakeit.Name(),
			Title:   gofakeit.Sentence(4),
			Summary: null.StringFrom(gofakeit.Sentence(10)),
			Body:    gofakeit.Sentence(20),
		})
}
