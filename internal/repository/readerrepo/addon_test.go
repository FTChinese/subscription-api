package readerrepo

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/invoice"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/test"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"testing"
	"time"
)

func TestEnv_ClaimAddOn(t *testing.T) {
	userID := uuid.New().String()

	t.Logf("User id: %s", userID)

	repo := test.NewRepo()
	repo.MustSaveMembership(reader.NewMockMemberBuilder(userID).
		WithExpiration(time.Now().AddDate(0, 0, -1)).
		Build())
	repo.MustSaveInvoiceN([]invoice.Invoice{
		reader.NewMockMemberBuilder(userID).
			Build().
			CarryOverInvoice(),
		invoice.NewMockInvoiceBuilder(userID).
			WithOrderKind(enum.OrderKindAddOn).
			Build(),
	})

	type fields struct {
		dbs    db.ReadWriteSplit
		logger *zap.Logger
	}
	type args struct {
		ids pkg.UserIDs
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    reader.AddOnClaimed
		wantErr bool
	}{
		{
			name: "Claim addon",
			fields: fields{
				dbs:    test.SplitDB,
				logger: zaptest.NewLogger(t),
			},
			args: args{
				ids: pkg.NewFtcUserID(userID),
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
			got, err := env.ClaimAddOn(tt.args.ids)
			if (err != nil) != tt.wantErr {
				t.Errorf("ClaimAddOn() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("ClaimAddOn() got = %v, want %v", got, tt.want)
			//}

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}
