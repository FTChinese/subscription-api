package addons

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/invoice"
	"github.com/FTChinese/subscription-api/test"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestEnv_InvoicesCarriedOver(t *testing.T) {
	userID := uuid.New().String()
	repo := test.NewRepo()
	repo.MustSaveInvoiceN([]invoice.Invoice{})

	env := Env{
		dbs:    test.SplitDB,
		logger: zaptest.NewLogger(t),
	}

	type args struct {
		userID ids.UserIDs
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Flag invoices as carried over",
			args: args{
				userID: ids.NewFtcUserID(userID),
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

func TestEnv_ListInvoices(t *testing.T) {
	ftcID := uuid.New().String()
	wxID := faker.WxUnionID()

	type fields struct {
		dbs    db.ReadWriteMyDBs
		logger *zap.Logger
	}
	type args struct {
		ids ids.UserIDs
		p   gorest.Pagination
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "List invoices",
			fields: fields{
				dbs:    test.SplitDB,
				logger: zaptest.NewLogger(t),
			},
			args: args{
				ids: ids.UserIDs{
					CompoundID: "",
					FtcID:      null.StringFrom(ftcID),
					UnionID:    null.StringFrom(wxID),
				}.MustNormalize(),
				p: gorest.NewPagination(1, 10),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				dbs:    tt.fields.dbs,
				logger: tt.fields.logger,
			}
			got, err := env.ListInvoices(tt.args.ids, tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListInvoices() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}
