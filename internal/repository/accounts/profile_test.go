package accounts

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/test"
	"github.com/brianvoe/gofakeit/v5"
	"github.com/guregu/null"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestEnv_LoadProfile(t *testing.T) {
	a := account.NewMockFtcAccountBuilder(enum.AccountKindFtc).Build()

	test.NewRepo().MustCreateFtcAccount(a)

	env := newTestEnv(db.MockMySQL(), zaptest.NewLogger(t))

	type args struct {
		ftcID string
	}
	tests := []struct {
		name    string
		args    args
		want    account.Profile
		wantErr bool
	}{
		{
			name: "Load profile",
			args: args{
				ftcID: a.FtcID,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := env.LoadProfile(tt.args.ftcID)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadProfile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("LoadProfile() got = %v, want %v", got, tt.want)
			//}

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}

func TestEnv_UpdateProfile(t *testing.T) {

	faker.SeedGoFake()

	a := account.NewMockFtcAccountBuilder(enum.AccountKindFtc).Build()

	test.NewRepo().MustCreateFtcAccount(a)

	t.Logf("%s", a.FtcID)

	env := newTestEnv(db.MockMySQL(), zaptest.NewLogger(t))

	type args struct {
		p account.BaseProfile
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Update profile",
			args: args{
				p: account.BaseProfile{
					ID:         a.FtcID,
					Gender:     enum.GenderMale,
					FamilyName: null.StringFrom(gofakeit.LastName()),
					GivenName:  null.StringFrom(gofakeit.FirstName()),
					Birthday:   null.StringFrom("1983-10-21"),
					CreatedUtc: chrono.Time{},
					UpdatedUtc: chrono.TimeNow(),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := env.UpdateProfile(tt.args.p); (err != nil) != tt.wantErr {
				t.Errorf("UpdateProfile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
