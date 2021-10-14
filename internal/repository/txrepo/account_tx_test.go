package txrepo

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/test"
	"github.com/guregu/null"
	"reflect"
	"testing"
)

func TestAccountTx_CreateAccount(t *testing.T) {
	a := account.NewMockFtcAccountBuilder(enum.AccountKindFtc).Build()

	type fields struct {
		SharedTx SharedTx
	}
	type args struct {
		a account.BaseAccount
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Create account",
			fields: fields{
				SharedTx: NewSharedTx(test.SplitDB.Write.MustBegin()),
			},
			args: args{
				a: a,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := AccountTx{
				SharedTx: tt.fields.SharedTx,
			}
			if err := tx.CreateAccount(tt.args.a); (err != nil) != tt.wantErr {
				t.Errorf("CreateAccount() error = %v, wantErr %v", err, tt.wantErr)
				_ = tx.Rollback()
				return
			}

			_ = tx.Commit()
		})
	}
}

func TestAccountTx_CreateProfile(t *testing.T) {
	a := account.NewMockFtcAccountBuilder(enum.AccountKindFtc).Build()

	type fields struct {
		SharedTx SharedTx
	}
	type args struct {
		a account.BaseAccount
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Create profile",
			fields: fields{
				SharedTx: NewSharedTx(test.SplitDB.Write.MustBegin()),
			},
			args: args{
				a: a,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := AccountTx{
				SharedTx: tt.fields.SharedTx,
			}
			if err := tx.CreateProfile(tt.args.a); (err != nil) != tt.wantErr {
				t.Errorf("CreateProfile() error = %v, wantErr %v", err, tt.wantErr)
				_ = tx.Rollback()
				return
			}

			_ = tx.Commit()
		})
	}
}

func TestAccountTx_AddUnionIDToFtc(t *testing.T) {
	a := account.NewMockFtcAccountBuilder(enum.AccountKindFtc).Build()

	t.Logf("ID: %s", a.FtcID)

	test.NewRepo().MustCreateFtcAccount(a)

	a.UnionID = null.StringFrom(faker.GenWxID())

	type fields struct {
		SharedTx SharedTx
	}
	type args struct {
		a account.BaseAccount
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Set union id to existing account",
			fields: fields{
				SharedTx: NewSharedTx(test.SplitDB.Write.MustBegin()),
			},
			args: args{
				a: a,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := AccountTx{
				SharedTx: tt.fields.SharedTx,
			}
			if err := tx.AddUnionIDToFtc(tt.args.a); (err != nil) != tt.wantErr {
				t.Errorf("AddUnionIDToFtc() error = %v, wantErr %v", err, tt.wantErr)
				_ = tx.Rollback()
				return
			}

			_ = tx.Commit()
		})
	}
}

func TestAccountTx_RetrieveMobiles(t *testing.T) {
	emailOnlyNoProfile := test.NewPersona().EmailOnlyAccount()
	emailOnlyWithProfile := test.NewPersona().EmailOnlyAccount()
	emailMobile := test.NewPersona().EmailMobileAccount()

	repo := test.NewRepo()

	repo.MustCreateUserInfo(emailOnlyNoProfile)
	repo.MustCreateFtcAccount(emailOnlyWithProfile)
	repo.MustCreateFtcAccount(emailMobile)

	type fields struct {
		SharedTx SharedTx
	}
	type args struct {
		u account.MobileUpdater
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []account.MobileUpdater
		wantErr bool
	}{
		{
			name: "Two rows from both side",
			fields: fields{
				SharedTx: NewSharedTx(db.MockMySQL().Read.MustBegin()),
			},
			args: args{
				u: account.MobileUpdater{
					FtcID:  emailOnlyWithProfile.FtcID,
					Mobile: emailMobile.Mobile,
				},
			},
			want: []account.MobileUpdater{
				{
					FtcID:  emailOnlyWithProfile.FtcID,
					Mobile: null.String{},
				},
				{
					FtcID:  emailMobile.FtcID,
					Mobile: emailMobile.Mobile,
				},
			},
			wantErr: false,
		},
		{
			name: "One row from mobile side",
			fields: fields{
				SharedTx: NewSharedTx(db.MockMySQL().Read.MustBegin()),
			},
			args: args{
				u: account.MobileUpdater{
					FtcID:  emailOnlyNoProfile.FtcID,
					Mobile: emailMobile.Mobile,
				},
			},
			want: []account.MobileUpdater{
				{
					FtcID:  emailMobile.FtcID,
					Mobile: emailMobile.Mobile,
				},
			},
			wantErr: false,
		},
		{
			name: "One row from ftc side",
			fields: fields{
				SharedTx: NewSharedTx(db.MockMySQL().Read.MustBegin()),
			},
			args: args{
				u: account.MobileUpdater{
					FtcID:  emailOnlyWithProfile.FtcID,
					Mobile: null.StringFrom(faker.GenPhone()),
				},
			},
			want: []account.MobileUpdater{
				{
					FtcID:  emailOnlyWithProfile.FtcID,
					Mobile: null.String{},
				},
			},
			wantErr: false,
		},
		{
			name: "No row",
			fields: fields{
				SharedTx: NewSharedTx(db.MockMySQL().Read.MustBegin()),
			},
			args: args{
				u: account.MobileUpdater{
					FtcID:  emailOnlyNoProfile.FtcID,
					Mobile: null.StringFrom(faker.GenPhone()),
				},
			},
			want:    []account.MobileUpdater{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := AccountTx{
				SharedTx: tt.fields.SharedTx,
			}
			got, err := tx.RetrieveMobiles(tt.args.u)
			if (err != nil) != tt.wantErr {
				t.Errorf("RetrieveMobiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RetrieveMobiles() got = %v, want %v", got, tt.want)
			}
		})
	}
}
