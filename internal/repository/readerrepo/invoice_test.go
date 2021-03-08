package readerrepo

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg"
	"github.com/FTChinese/subscription-api/pkg/invoice"
	"github.com/FTChinese/subscription-api/test"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"testing"
	"time"
)

func TestEnv_InvoicesCarriedOver(t *testing.T) {
	userID := uuid.New().String()
	repo := test.NewRepo()
	repo.MustSaveInvoiceN([]invoice.Invoice{
		invoice.NewMockInvoiceBuilder(userID).
			Build().
			SetPeriod(time.Now()),
		invoice.NewMockInvoiceBuilder(userID).
			WithOrderKind(enum.OrderKindRenew).
			Build().
			SetPeriod(time.Now().AddDate(1, 0, 1)),
	})
	type fields struct {
		db     *sqlx.DB
		logger *zap.Logger
	}
	type args struct {
		userID pkg.UserIDs
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Flag invoices as carried over",
			fields: fields{
				db:     test.DB,
				logger: zaptest.NewLogger(t),
			},
			args: args{
				userID: pkg.NewFtcUserID(userID),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				db:     tt.fields.db,
				logger: tt.fields.logger,
			}
			if err := env.InvoicesCarriedOver(tt.args.userID); (err != nil) != tt.wantErr {
				t.Errorf("InvoicesCarriedOver() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
