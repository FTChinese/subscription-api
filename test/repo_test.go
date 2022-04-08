package test

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/pw"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"testing"
)

func TestRepo_SaveMembership(t *testing.T) {
	p := NewPersona()

	err := NewRepo().SaveMembership(p.MemberBuilder().WithPrice(pw.MockPwPriceStdYear.FtcPrice).Build())

	if err != nil {
		t.Error(err)
	}
}

func TestRepo_SaveOrder(t *testing.T) {
	p := NewPersona()

	err := NewRepo().SaveOrder(p.OrderBuilder().Build())

	if err != nil {
		t.Error(err)
	}
}

func TestRepo_SaveIAPSubs(t *testing.T) {
	p := NewPersona()

	err := NewRepo().SaveIAPSubs(p.IAPBuilder().Build())
	if err != nil {
		t.Error(err)
	}
}

// Generate a mocking wx user
func TestRepo_SaveWxUser(t *testing.T) {
	unionID := faker.GenWxID()
	t.Logf("Wx user %s", unionID)

	schema := NewPersona().WxUser()
	err := NewRepo().SaveWxUser(schema)
	if err != nil {
		t.Error(err)
	}
}

func TestRepo_CreateWxUser(t *testing.T) {
	p := NewPersona()

	repo := NewRepo()
	w := p.WxUser()
	t.Logf("%+v", w)
	err := repo.SaveWxUser(w)
	if err != nil {
		t.Error(err)
		return
	}

	m := reader.NewMockMemberBuilderV2(enum.AccountKindWx).
		WithWxID(p.UnionID).
		Build()
	t.Logf("%+v", m)
	repo.MustSaveMembership(m)
}

func TestRepo_WxWebhook(t *testing.T) {
	p := NewNPC()

	o := p.OrderBuilder().Build()

	err := NewRepo().SaveOrder(o)

	if err != nil {
		t.Error(err)
	}

	payload := NewWxWebhookPayload(o)

	t.Logf("\n%s\n", payload.ToXML())
}
