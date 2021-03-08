package readerrepo

import (
	"github.com/FTChinese/subscription-api/pkg"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/test"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestReaderEnv_AccountByFtcID(t *testing.T) {
	profile := test.NewPersona()
	test.NewRepo().MustSaveAccount(profile.FtcAccount())

	env := Env{db: test.DB}

	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "Find account ftc id",
			args:    args{id: profile.FtcID},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := env.AccountByFtcID(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindAccountByFtcID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("Account: %+v", got)
		})
	}
}

func TestEnv_AccountByWxID(t *testing.T) {
	p := test.NewPersona()

	repo := test.NewRepo()
	repo.MustSaveWxUser(p.WxUser())

	type fields struct {
		db *sqlx.DB
	}
	type args struct {
		unionID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    reader.FtcAccount
		wantErr bool
	}{
		{
			name: "Wx account",
			fields: fields{
				db: test.DB,
			},
			args: args{
				unionID: p.UnionID,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				db: tt.fields.db,
			}
			got, err := env.AccountByWxID(tt.args.unionID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("AccountByWxID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("AccountByWxID() got = %v, want %v", got, tt.want)
			//}
			assert.NotZero(t, got)
		})
	}
}

func TestReaderEnv_AccountByStripeID(t *testing.T) {
	profile := test.NewPersona()
	test.NewRepo().MustSaveAccount(profile.FtcAccount())

	env := Env{db: test.DB}

	type args struct {
		cusID string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "Account by stripe",
			args:    args{cusID: profile.StripeID},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := env.FtcAccountByStripeID(tt.args.cusID)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindAccountByStripeID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("Acocunt: %+v", got)
		})
	}
}

func TestEnv_RetrieveMember(t *testing.T) {
	p := test.NewPersona()
	m := p.Membership()
	test.NewRepo().MustSaveMembership(m)

	env := NewEnv(test.DB, zaptest.NewLogger(t))

	type args struct {
		id pkg.UserIDs
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Load member",
			args: args{
				id: m.UserIDs,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := env.RetrieveMember(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("RetrieveMember() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("%v", got)
		})
	}
}
