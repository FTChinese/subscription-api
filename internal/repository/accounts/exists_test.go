package accounts

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/internal/repository/readers"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/test"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestEnv_IDExists(t *testing.T) {

	env := New(test.SplitDB, zaptest.NewLogger(t))

	a := account.NewMockFtcAccountBuilder(enum.AccountKindFtc).Build()

	test.NewRepo().MustCreateFtcAccount(a)

	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "",
			args: args{
				id: a.FtcID,
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := env.IDExists(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("IDExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IDExists() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnv_EmailExists(t *testing.T) {

	a := account.NewMockFtcAccountBuilder(enum.AccountKindFtc).Build()
	t.Logf("%s : %s", a.FtcID, a.Email)

	test.NewRepo().MustCreateFtcAccount(a)

	type fields struct {
		Env readers.Env
	}
	type args struct {
		email string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "Email exists",
			fields: fields{
				Env: readers.New(test.SplitDB, zaptest.NewLogger(t)),
			},
			args: args{
				email: a.Email,
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				Env: tt.fields.Env,
			}
			got, err := env.EmailExists(tt.args.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("EmailExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("EmailExists() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnv_NameExists(t *testing.T) {
	a := account.NewMockFtcAccountBuilder(enum.AccountKindFtc).Build()
	t.Logf("%s : %s", a.FtcID, a.Email)

	test.NewRepo().MustCreateFtcAccount(a)

	type fields struct {
		Env readers.Env
	}
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "Name exists",
			fields: fields{
				Env: readers.New(test.SplitDB, zaptest.NewLogger(t)),
			},
			args: args{
				name: a.UserName.String,
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				Env: tt.fields.Env,
			}
			got, err := env.NameExists(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("NameExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("NameExists() got = %v, want %v", got, tt.want)
			}
		})
	}
}
