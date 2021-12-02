package accounts

import (
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/test"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestEnv_JoinedByFtcID(t *testing.T) {
	p := test.NewPersona()

	a := p.EmailWxAccount()
	w := p.WxUser()

	repo := test.NewRepo()
	repo.MustCreateFtcAccount(a)
	repo.MustSaveWxUser(w)

	env := newTestEnv(db.MockMySQL(), zaptest.NewLogger(t))

	type args struct {
		ftcID string
	}
	tests := []struct {
		name    string
		args    args
		want    account.JoinedSchema
		wantErr bool
	}{
		{
			name: "Retrieve joined accont of ftc and wechat",
			args: args{
				ftcID: a.FtcID,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
	p := test.NewPersona()

	a := p.EmailWxAccount()
	w := p.WxUser()

	repo := test.NewRepo()
	repo.MustCreateFtcAccount(a)
	repo.MustSaveWxUser(w)

	env := newTestEnv(db.MockMySQL(), zaptest.NewLogger(t))

	type args struct {
		unionID string
	}
	tests := []struct {
		name    string
		args    args
		want    account.JoinedSchema
		wantErr bool
	}{
		{
			name: "Retrieve account by wechat id",
			args: args{
				unionID: w.UnionID,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
