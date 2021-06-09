// +build !production

package test

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/ztsms"
	"github.com/guregu/null"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestMockBaseAccount(t *testing.T) {
	a := account.
		NewMockFtcAccountBuilder(enum.AccountKindLinked).
		Build()

	t.Logf("%s", faker.MustMarshalIndent(a))
}

func TestLinkMobile(t *testing.T) {
	a := account.NewMockFtcAccountBuilder(enum.AccountKindFtc).WithMobile("").Build()
	v := ztsms.NewVerifier(faker.GenPhone(), null.String{})

	repo := NewRepoV2(zaptest.NewLogger(t))
	repo.MustCreateFtcAccount(a)
	repo.MustSaveMobileVerifier(v)

	t.Logf("%s", faker.MustMarshalIndent(a))
	t.Logf("%s", faker.MustMarshalIndent(v))
}

func TestUnsetMobile(t *testing.T) {
	a := account.NewMockFtcAccountBuilder(enum.AccountKindFtc).Build()

	NewRepo().MustCreateFtcAccount(a)

	t.Logf("%s", faker.MustMarshalIndent(a))
}
