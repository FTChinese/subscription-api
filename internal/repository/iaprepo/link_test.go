package iaprepo

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/apple"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/test"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEnv_Link(t *testing.T) {

	p1 := test.NewPersona()
	p2 := test.NewPersona().SetExpired(true)
	p3 := test.NewPersona().SetPayMethod(enum.PayMethodApple)

	repo := test.NewRepo()
	repo.MustSaveMembership(p1.Membership())
	repo.MustSaveMembership(p2.Membership())
	repo.MustSaveMembership(p3.Membership())

	type fields struct {
		cfg config.BuildConfig
		db  *sqlx.DB
	}
	type args struct {
		account reader.FtcAccount
		iapSubs apple.Subscription
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
				db: test.DB,
			},
			args: args{
				account: test.NewPersona().FtcAccount(),
				iapSubs: test.NewPersona().IAPSubs(),
			},
			wantErr: false,
		},
		{
			name: "IAP current empty cannot link to FTC non-expired",
			fields: fields{
				cfg: config.BuildConfig{},
				db:  test.DB,
			},
			args: args{
				account: p1.FtcAccount(),
				iapSubs: p1.IAPSubs(),
			},
			wantErr: true,
		},
		{
			name: "IAP current empty can link to FTC expired",
			fields: fields{
				cfg: config.BuildConfig{},
				db:  test.DB,
			},
			args: args{
				account: p2.FtcAccount(),
				iapSubs: test.NewPersona().IAPSubs(),
			},
			wantErr: false,
		},
		{
			name: "IAP exists and should not link to any new ftc account",
			fields: fields{
				db: test.DB,
			},
			args: args{
				account: test.NewPersona().FtcAccount(),
				iapSubs: p3.IAPSubs(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				db: tt.fields.db,
			}
			got, err := env.Link(tt.args.account, tt.args.iapSubs)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

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
