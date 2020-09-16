package subrepo

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/FTChinese/subscription-api/test"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestEnv_CreateOrder(t *testing.T) {

	type fields struct {
		db     *sqlx.DB
		logger *zap.Logger
	}
	type args struct {
		builder *subs.OrderBuilder
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Wx order",
			fields: fields{
				db:     test.DB,
				logger: zaptest.NewLogger(t),
			},
			args: args{
				builder: test.NewPersona().
					SetPayMethod(enum.PayMethodWx).
					WxOrderBuilder(),
			},
			wantErr: false,
		},
		{
			name: "Ali order",
			fields: fields{
				db:     test.DB,
				logger: zaptest.NewLogger(t),
			},
			args: args{
				builder: test.NewPersona().
					AliOrderBuilder(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				db:     tt.fields.db,
				logger: tt.fields.logger,
			}
			got, err := env.CreateOrder(tt.args.builder)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}
