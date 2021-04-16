package accounts

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/internal/repository/readers"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/wxlogin"
	"github.com/FTChinese/subscription-api/test"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestEnv_JoinedByFtcID(t *testing.T) {
	a := account.NewMockFtcAccountBuilder(enum.AccountKindLinked).Build()
	w := wxlogin.MockUserInfo(a.UnionID.String)

	repo := test.NewRepo()
	repo.MustCreateFtcAccount(a)
	repo.MustSaveWxUser(w)

	type fields struct {
		Env readers.Env
	}
	type args struct {
		ftcID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    account.JoinedSchema
		wantErr bool
	}{
		{
			name: "Retrieve joined accont of ftc and wechat",
			fields: fields{
				Env: readers.New(test.SplitDB, zaptest.NewLogger(t)),
			},
			args: args{
				ftcID: a.FtcID,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				Env: tt.fields.Env,
			}
			got, err := env.JoinedByFtcID(tt.args.ftcID)
			if (err != nil) != tt.wantErr {
				t.Errorf("JoinedByFtcID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("JoinedByFtcID() got = %v, want %v", got, tt.want)
			//}

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}

func TestEnv_JoinedByWxID(t *testing.T) {
	a := account.NewMockFtcAccountBuilder(enum.AccountKindLinked).Build()
	w := wxlogin.MockUserInfo(a.UnionID.String)

	repo := test.NewRepo()
	repo.MustCreateFtcAccount(a)
	repo.MustSaveWxUser(w)

	type fields struct {
		Env readers.Env
	}
	type args struct {
		unionID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    account.JoinedSchema
		wantErr bool
	}{
		{
			name: "Retrieve account by wechat id",
			fields: fields{
				Env: readers.New(test.SplitDB, zaptest.NewLogger(t)),
			},
			args: args{
				unionID: w.UnionID,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				Env: tt.fields.Env,
			}
			got, err := env.JoinedByWxID(tt.args.unionID)
			if (err != nil) != tt.wantErr {
				t.Errorf("JoinedByWxID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("JoinedByWxID() got = %v, want %v", got, tt.want)
			//}

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}
