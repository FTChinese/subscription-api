package subrepo

import (
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/FTChinese/subscription-api/test"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEnv_ConfirmOrder(t *testing.T) {
	repo := test.NewRepo()

	p1 := test.NewPersona()
	order1 := p1.CreateOrder()
	t.Logf("Ali Order id %s", order1.ID)
	repo.MustSaveOrder(order1)

	p2 := test.NewPersona()
	order2 := p2.CreateOrder()
	t.Logf("Wx Order id %s", order2.ID)
	repo.MustSaveOrder(order2)

	type fields struct {
		db *sqlx.DB
	}
	type args struct {
		result subs.PaymentResult
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Confirm ali order",
			fields: fields{
				db: test.DB,
			},
			args: args{
				result: p1.PaymentResult(order1),
			},
			wantErr: false,
		},
		{
			name: "Confirm wx order",
			fields: fields{
				db: test.DB,
			},
			args: args{
				result: p1.PaymentResult(order2),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				db: tt.fields.db,
			}

			t.Logf("Payment result: %+v", tt.args.result)

			got, err := env.ConfirmOrder(tt.args.result)
			if tt.wantErr {
				assert.NotNil(t, err)
				return
			}

			assert.Nil(t, err)
			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}
