//go:build !production

package android

import (
	"github.com/FTChinese/subscription-api/faker"
	"github.com/brianvoe/gofakeit/v5"
	"github.com/guregu/null"
)

func NewMockRelease() Release {
	faker.SeedGoFake()

	return NewRelease(ReleaseInput{
		VersionName: gofakeit.AppVersion(),
		VersionCode: int64(gofakeit.Uint16()),
		Body:        null.StringFrom(gofakeit.Sentence(20)),
		ApkURL:      gofakeit.URL(),
	})
}
