package test

import (
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewProfile(t *testing.T) {
	t.Log(NewPersona().Email)
	t.Log(NewPersona().Email)
}

func TestPersona_ConfirmOrder(t *testing.T) {
	p := NewPersona()
	o := p.CreateOrder()

	result := p.ConfirmOrder(o)

	assert.NotZero(t, result.Order.ConfirmedAt)
	assert.NotZero(t, result.Order.StartDate)
	assert.NotZero(t, result.Order.EndDate)
}

func TestPersona_IAPSubs(t *testing.T) {
	p := NewPersona()

	sub := p.IAPSubs()

	m := sub.NewMembership(p.AccountID())

	m = m.Sync()

	assert.NotZero(t, m.LegacyExpire)
	assert.NotZero(t, m.LegacyTier)

	t.Log(m.LegacyTier)
	t.Log(m.LegacyExpire)
}

func TestPersona_PaymentResult(t *testing.T) {
	p1 := NewPersona()
	p2 := NewPersona()

	type args struct {
		order subs.Order
	}
	tests := []struct {
		name   string
		fields *Persona
		args   args
	}{
		{
			name:   "Alipay result",
			fields: p1,
			args: args{
				order: p1.CreateOrder(),
			},
		},
		{
			name:   "Wxpay result",
			fields: p2,
			args: args{
				order: p2.CreateOrder(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fields.PaymentResult(tt.args.order)

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}
