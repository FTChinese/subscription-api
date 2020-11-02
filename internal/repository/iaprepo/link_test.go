package iaprepo

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/apple"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/test"
	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestEnv_GetSubAndSetFtcID(t *testing.T) {
	p := test.NewPersona().SetPayMethod(enum.PayMethodApple)

	test.NewRepo().MustSaveIAPSubs(p.IAPSubs())

	type fields struct {
		db     *sqlx.DB
		rdb    *redis.Client
		logger *zap.Logger
	}
	type args struct {
		input apple.LinkInput
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Get subscription and optionally set ftc id",
			fields: fields{
				db:     test.DB,
				logger: zaptest.NewLogger(t),
			},
			args: args{
				input: apple.LinkInput{
					FtcID:        p.FtcID,
					OriginalTxID: p.AppleSubID,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				db:     tt.fields.db,
				rdb:    tt.fields.rdb,
				logger: tt.fields.logger,
			}
			got, err := env.GetSubAndSetFtcID(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSubAndSetFtcID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("%+v", got)
		})
	}
}

func TestEnv_Link(t *testing.T) {

	p1 := test.NewPersona()
	p2 := test.NewPersona().SetExpired(true)
	p3 := test.NewPersona().SetPayMethod(enum.PayMethodApple)

	repo := test.NewRepo()
	repo.MustSaveMembership(p1.Membership())
	repo.MustSaveMembership(p2.Membership())
	repo.MustSaveMembership(p3.Membership())

	type fields struct {
		db     *sqlx.DB
		rdb    *redis.Client
		logger *zap.Logger
	}
	type args struct {
		account reader.FtcAccount
		sub     apple.Subscription
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Both sides have no existing membership",
			fields: fields{
				db:     test.DB,
				rdb:    test.Redis,
				logger: zaptest.NewLogger(t),
			},
			args: args{
				account: test.NewPersona().FtcAccount(),
				sub:     test.NewPersona().IAPSubs(),
			},
			wantErr: false,
		},
		{
			name: "IAP current empty cannot link to FTC non-expired",
			fields: fields{
				db:     test.DB,
				rdb:    test.Redis,
				logger: zaptest.NewLogger(t),
			},
			args: args{
				account: p1.FtcAccount(),
				sub:     p1.IAPSubs(),
			},
			wantErr: true,
		},
		{
			name: "IAP current empty can link to FTC expired",
			fields: fields{
				db:     test.DB,
				rdb:    test.Redis,
				logger: zaptest.NewLogger(t),
			},
			args: args{
				account: p2.FtcAccount(),
				sub:     test.NewPersona().IAPSubs(),
			},
			wantErr: false,
		},
		{
			name: "IAP exists and should not link to any new ftc account",
			fields: fields{
				db:     test.DB,
				rdb:    test.Redis,
				logger: zaptest.NewLogger(t),
			},
			args: args{
				account: test.NewPersona().FtcAccount(),
				sub:     p3.IAPSubs(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				db:     tt.fields.db,
				rdb:    tt.fields.rdb,
				logger: tt.fields.logger,
			}
			got, err := env.Link(tt.args.account, tt.args.sub)
			if (err != nil) != tt.wantErr {
				t.Errorf("Link() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}

func TestEnv_Unlink(t *testing.T) {

	p := test.NewPersona().SetPayMethod(enum.PayMethodApple)

	test.NewRepo().MustSaveMembership(p.Membership())

	type fields struct {
		cfg config.BuildConfig
		db  *sqlx.DB
	}
	type args struct {
		input apple.LinkInput
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    reader.MemberSnapshot
		wantErr bool
	}{
		{
			name: "Unlink IAP",
			fields: fields{
				db: test.DB,
			},
			args: args{
				input: apple.LinkInput{
					FtcID:        p.FtcID,
					OriginalTxID: p.AppleSubID,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				db: tt.fields.db,
			}
			got, err := env.Unlink(tt.args.input)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}
