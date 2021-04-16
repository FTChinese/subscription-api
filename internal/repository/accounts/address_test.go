package accounts

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/internal/repository/readers"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/test"
	"github.com/brianvoe/gofakeit/v5"
	"github.com/guregu/null"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestEnv_LoadAddress(t *testing.T) {
	a := account.NewMockFtcAccountBuilder(enum.AccountKindFtc).Build()
	test.NewRepo().MustCreateFtcAccount(a)

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
		want    account.Address
		wantErr bool
	}{
		{
			name: "Load address",
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
			got, err := env.LoadAddress(tt.args.ftcID)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("LoadAddress() got = %v, want %v", got, tt.want)
			//}

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}

func TestEnv_UpdateAddress(t *testing.T) {
	faker.SeedGoFake()

	a := account.NewMockFtcAccountBuilder(enum.AccountKindFtc).Build()

	t.Logf("%s", a.FtcID)

	test.NewRepo().MustCreateFtcAccount(a)

	type fields struct {
		Env readers.Env
	}
	type args struct {
		addr account.Address
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Update address",
			fields: fields{
				Env: readers.New(test.SplitDB, zaptest.NewLogger(t)),
			},
			args: args{
				addr: account.Address{
					FtcID:    a.FtcID,
					Country:  null.StringFrom(gofakeit.Country()),
					Province: null.StringFrom(gofakeit.State()),
					City:     null.StringFrom(gofakeit.City()),
					District: null.StringFrom(gofakeit.City()),
					Street:   null.StringFrom(gofakeit.Street()),
					Postcode: null.StringFrom(gofakeit.Zip()),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				Env: tt.fields.Env,
			}
			if err := env.UpdateAddress(tt.args.addr); (err != nil) != tt.wantErr {
				t.Errorf("UpdateAddress() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
