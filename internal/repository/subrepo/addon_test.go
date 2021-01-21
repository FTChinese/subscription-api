package subrepo

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/test"
	"github.com/guregu/null"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"reflect"
	"testing"
	"time"
)

func TestEnv_RedeemAddOn(t *testing.T) {
	p := test.NewPersona().
		SetPayMethod(enum.PayMethodStripe).
		SetExpired(true).
		SetAutoRenew(false).
		SetReservedDays(reader.ReservedDays{
			Standard: 366*3 + 3,
			Premium:  0,
		})

	m := p.Membership()
	aos := p.AddOnN(3)

	repo := test.NewRepo()
	repo.MustSaveMembership(m)
	repo.MustSaveAddOnN(aos)

	type fields struct {
		rwdDB  *sqlx.DB
		logger *zap.Logger
	}
	type args struct {
		ids reader.MemberID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    reader.Membership
		wantErr bool
	}{
		{
			name: "Redeem add-ons",
			fields: fields{
				rwdDB:  test.DB,
				logger: zaptest.NewLogger(t),
			},
			args: args{
				ids: p.AccountID(),
			},
			want: reader.Membership{
				MemberID:      p.AccountID(),
				Edition:       aos[2].Edition,
				LegacyTier:    null.Int{},
				LegacyExpire:  null.Int{},
				ExpireDate:    chrono.DateFrom(time.Now().AddDate(3, 0, 3)),
				PaymentMethod: aos[2].PaymentMethod,
				FtcPlanID:     aos[2].PlanID,
				ReservedDays:  reader.ReservedDays{},
			}.Sync(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				rwdDB:  tt.fields.rwdDB,
				logger: tt.fields.logger,
			}
			got, err := env.RedeemAddOn(tt.args.ids)
			if (err != nil) != tt.wantErr {
				t.Errorf("RedeemAddOn() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.Membership, tt.want) {
				t.Errorf("RedeemAddOn() got = %v\n, want %v", got.Membership, tt.want)
			}

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}
