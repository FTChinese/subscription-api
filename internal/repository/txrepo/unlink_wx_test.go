package txrepo

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/test"
	"github.com/jmoiron/sqlx"
	"testing"
)

func TestUnlinkTx_UnlinkUser(t *testing.T) {
	a := account.NewMockFtcAccountBuilder(enum.AccountKindLinked).Build()
	test.NewRepo().MustCreateFtcAccount(a)

	type fields struct {
		Tx *sqlx.Tx
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
			name: "Unlink account",
			fields: fields{
				Tx: test.SplitDB.Write.MustBegin(),
			},
			args: args{
				a: a,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := UnlinkTx{
				Tx: tt.fields.Tx,
			}
			if err := tx.UnlinkUser(tt.args.a); (err != nil) != tt.wantErr {
				t.Errorf("UnlinkUser() error = %v, wantErr %v", err, tt.wantErr)
				_ = tx.Rollback()
				return
			}

			_ = tx.Commit()
		})
	}
}

func TestUnlinkTx_dropWxFromMember(t *testing.T) {
	m := reader.NewMockMemberBuilderV2(enum.AccountKindLinked).Build()

	test.NewRepo().MustSaveMembership(m)

	type fields struct {
		Tx *sqlx.Tx
	}
	type args struct {
		m reader.Membership
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Drop wx id from membership",
			fields: fields{
				Tx: test.SplitDB.Write.MustBegin(),
			},
			args: args{
				m: m,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := UnlinkTx{
				Tx: tt.fields.Tx,
			}
			if err := tx.dropWxFromMember(tt.args.m); (err != nil) != tt.wantErr {
				t.Errorf("dropWxFromMember() error = %v, wantErr %v", err, tt.wantErr)
				_ = tx.Rollback()
				return
			}

			_ = tx.Commit()
		})
	}
}

func TestUnlinkTx_dropFtcFromMember(t *testing.T) {
	m := reader.NewMockMemberBuilderV2(enum.AccountKindLinked).Build()

	test.NewRepo().MustSaveMembership(m)

	type fields struct {
		Tx *sqlx.Tx
	}
	type args struct {
		m reader.Membership
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Drop ftc id from membership",
			fields: fields{
				Tx: test.SplitDB.Write.MustBegin(),
			},
			args: args{
				m: m,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := UnlinkTx{
				Tx: tt.fields.Tx,
			}
			if err := tx.dropFtcFromMember(tt.args.m); (err != nil) != tt.wantErr {
				t.Errorf("dropFtcFromMember() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
