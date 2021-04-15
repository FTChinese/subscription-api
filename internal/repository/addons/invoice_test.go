package addons

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg"
	"github.com/FTChinese/subscription-api/pkg/invoice"
	"github.com/FTChinese/subscription-api/test"
	"github.com/google/uuid"
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

	env := Env{
		dbs:    test.SplitDB,
		logger: zaptest.NewLogger(t),
	}

	type args struct {
		userID pkg.UserIDs
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Flag invoices as carried over",
			args: args{
				userID: pkg.NewFtcUserID(userID),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := env.InvoicesCarriedOver(tt.args.userID); (err != nil) != tt.wantErr {
				t.Errorf("InvoicesCarriedOver() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
