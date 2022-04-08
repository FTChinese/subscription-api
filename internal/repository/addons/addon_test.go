package addons

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/addon"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/ids"
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
	// User A has valid and complete invoices.
	userA := uuid.New().String()
	t.Logf("User A: %s", userA)

	// User B only has addon days under membership.
	userB := uuid.New().String()
	t.Logf("User B: %s", userB)

	repo := test.NewRepo()

	type requisite struct {
		membership reader.Membership
		invoices   []invoice.Invoice
	}

	type fields struct {
		dbs    db.ReadWriteMyDBs
		logger *zap.Logger
	}
	type args struct {
		ids ids.UserIDs
	}
	tests := []struct {
		name      string
		fields    fields
		requisite requisite
		args      args
		want      reader.AddOnClaimed
		wantErr   bool
	}{
		{
			name: "Claim addon from invoices",
			fields: fields{
				dbs:    test.SplitDB,
				logger: zaptest.NewLogger(t),
			},
			requisite: requisite{
				membership: reader.NewMockMemberBuilderV2(enum.AccountKindFtc).
					WithFtcID(userA).
					WithExpiration(time.Now().AddDate(0, 0, -1)).
					Build(),
				invoices: []invoice.Invoice{
					reader.NewMockMemberBuilderV2(enum.AccountKindFtc).
						WithFtcID(userA).
						Build().
						CarryOverInvoice(),
					invoice.Invoice{},
				},
			},
			args: args{
				ids: ids.NewFtcUserID(userA),
			},
			wantErr: false,
		},
		{
			name: "Claim addon from membership days",
			fields: fields{
				dbs:    test.SplitDB,
				logger: zaptest.NewLogger(t),
			},
			requisite: requisite{
				membership: reader.NewMockMemberBuilderV2(enum.AccountKindFtc).
					WithFtcID(userB).
					WithExpiration(time.Now().AddDate(0, 0, -1)).
					WithAddOn(addon.AddOn{
						Standard: 367,
						Premium:  0,
					}).
					Build(),
				invoices: nil,
			},
			args: args{
				ids: ids.NewFtcUserID(userB),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Save membership %v", tt.requisite.membership)
			repo.MustSaveMembership(tt.requisite.membership)
			t.Logf("Save invoices %v", tt.requisite.invoices)
			repo.MustSaveInvoiceN(tt.requisite.invoices)

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
