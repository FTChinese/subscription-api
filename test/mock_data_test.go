package test

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/account"
	"testing"
)

func TestMockBaseAccount(t *testing.T) {
	a := account.NewMockFtcAccountBuilder(enum.AccountKindLinked).Build()

	t.Logf("%s", faker.MustMarshalIndent(a))
}
