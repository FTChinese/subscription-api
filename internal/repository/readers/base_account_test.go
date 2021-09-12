package readers

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/test"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestEnv_BaseAccountByUUID(t *testing.T) {
	a := account.NewMockFtcAccountBuilder(enum.AccountKindFtc).Build()

	repo := test.NewRepo()
	repo.MustCreateFtcAccount(a)

	type fields struct {
		DBs    db.ReadWriteMyDBs
		Logger *zap.Logger
	}
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    account.BaseAccount
		wantErr bool
	}{
		{
			name: "Retrieve account by uuid",
			fields: fields{
				DBs:    test.SplitDB,
				Logger: zaptest.NewLogger(t),
			},
			args: args{
				id: a.FtcID,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				DBs:    tt.fields.DBs,
				Logger: tt.fields.Logger,
			}
			got, err := env.BaseAccountByUUID(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("BaseAccountByUUID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("BaseAccountByUUID() got = %v, want %v", got, tt.want)
			//}

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}

func TestEnv_BaseAccountByStripeID(t *testing.T) {
	a := account.NewMockFtcAccountBuilder(enum.AccountKindFtc).Build()

	test.NewRepo().MustCreateFtcAccount(a)

	type fields struct {
		DBs    db.ReadWriteMyDBs
		Logger *zap.Logger
	}
	type args struct {
		cusID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    account.BaseAccount
		wantErr bool
	}{
		{
			name: "Retrieve account by stripe customer id",
			fields: fields{
				DBs:    test.SplitDB,
				Logger: zaptest.NewLogger(t),
			},
			args: args{
				cusID: a.StripeID.String,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				DBs:    tt.fields.DBs,
				Logger: tt.fields.Logger,
			}
			got, err := env.BaseAccountByStripeID(tt.args.cusID)
			if (err != nil) != tt.wantErr {
				t.Errorf("BaseAccountByStripeID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("BaseAccountByStripeID() got = %v, want %v", got, tt.want)
			//}

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}

func TestEnv_BaseAccountByWxID(t *testing.T) {
	a := account.NewMockFtcAccountBuilder(enum.AccountKindLinked).Build()
	test.NewRepo().MustCreateFtcAccount(a)

	type fields struct {
		DBs    db.ReadWriteMyDBs
		Logger *zap.Logger
	}
	type args struct {
		unionID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    account.BaseAccount
		wantErr bool
	}{
		{
			name: "Retrieve account by wechat id",
			fields: fields{
				DBs:    test.SplitDB,
				Logger: zaptest.NewLogger(t),
			},
			args: args{
				unionID: a.UnionID.String,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				DBs:    tt.fields.DBs,
				Logger: tt.fields.Logger,
			}
			got, err := env.BaseAccountByWxID(tt.args.unionID)
			if (err != nil) != tt.wantErr {
				t.Errorf("BaseAccountByWxID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("BaseAccountByWxID() got = %v, want %v", got, tt.want)
			//}

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}
