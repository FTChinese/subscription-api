package iaprepo

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/apple"
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

func TestEnv_ArchiveLinkCheating(t *testing.T) {

	p := test.NewPersona()

	type fields struct {
		db     *sqlx.DB
		rdb    *redis.Client
		logger *zap.Logger
	}
	type args struct {
		link apple.LinkInput
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Archive link cheating",
			fields: fields{
				db:     test.DB,
				rdb:    test.Redis,
				logger: zaptest.NewLogger(t),
			},
			args: args{link: apple.LinkInput{
				FtcID:        p.FtcID,
				OriginalTxID: p.AppleSubID,
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				db:     tt.fields.db,
				rdb:    tt.fields.rdb,
				logger: tt.fields.logger,
			}
			if err := env.ArchiveLinkCheating(tt.args.link); (err != nil) != tt.wantErr {
				t.Errorf("ArchiveLinkCheating() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnv_Unlink(t *testing.T) {

	p := test.NewPersona().SetPayMethod(enum.PayMethodApple)

	sub := p.IAPSubs()

	repo := test.NewRepo()
	repo.MustSaveMembership(p.Membership())
	repo.MustSaveIAPSubs(sub)

	env := NewEnv(test.DB, test.Redis, zaptest.NewLogger(t))

	type args struct {
		input apple.LinkInput
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Unlink IAP",
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

func TestEnv_ArchiveUnlink(t *testing.T) {

	p := test.NewPersona()

	env := NewEnv(test.DB, test.Redis, zaptest.NewLogger(t))

	type args struct {
		link apple.LinkInput
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Archive unlink",
			args: args{
				link: apple.LinkInput{
					FtcID:        p.FtcID,
					OriginalTxID: p.AppleSubID,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := env.ArchiveUnlink(tt.args.link); (err != nil) != tt.wantErr {
				t.Errorf("ArchiveUnlink() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
