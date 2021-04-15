package txrepo

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/test"
	"github.com/guregu/null"
	"github.com/jmoiron/sqlx"
	"testing"
)

func TestStripeTx_BaseAccountForStripe(t *testing.T) {

	a := account.NewMockFtcAccountBuilder(enum.AccountKindFtc).Build()

	test.NewRepo().MustCreateFtcAccount(a)

	type fields struct {
		Tx *sqlx.Tx
	}
	type args struct {
		ftcID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    account.BaseAccount
		wantErr bool
	}{
		{
			name: "Retrieve base account",
			fields: fields{
				Tx: test.SplitDB.Read.MustBegin(),
			},
			args: args{
				ftcID: a.FtcID,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := NewStripeTx(tt.fields.Tx)
			got, err := tx.BaseAccountForStripe(tt.args.ftcID)
			if (err != nil) != tt.wantErr {
				t.Errorf("BaseAccountForStripe() error = %v, wantErr %v", err, tt.wantErr)
				_ = tx.Rollback()
				return
			}

			_ = tx.Commit()

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}

func TestStripeTx_SaveCustomerID(t *testing.T) {

	a := account.NewMockFtcAccountBuilder(enum.AccountKindFtc).Build()

	t.Logf("Creating account %s", a.FtcID)

	test.NewRepo().MustCreateFtcAccount(a)

	a.StripeID = null.StringFrom(faker.GenCustomerID())

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
			name: "Set stripe customer id",
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
			tx := NewStripeTx(tt.fields.Tx)
			if err := tx.SaveCustomerID(tt.args.a); (err != nil) != tt.wantErr {
				t.Errorf("SaveCustomerID() error = %v, wantErr %v", err, tt.wantErr)
				_ = tx.Rollback()
				return
			}

			_ = tx.Commit()
		})
	}
}

func TestStripeTx_RetrieveStripeMember(t *testing.T) {

	m := reader.NewMockMemberBuilder("").
		WithPayMethod(enum.PayMethodStripe).
		Build()

	test.NewRepo().MustSaveMembership(m)

	type fields struct {
		Tx *sqlx.Tx
	}

	type args struct {
		subID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Retrieve stripe member",
			fields: fields{
				Tx: test.SplitDB.Read.MustBegin(),
			},
			args: args{
				subID: m.StripeSubsID.String,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := NewStripeTx(tt.fields.Tx)

			got, err := tx.RetrieveStripeMember(tt.args.subID)
			if (err != nil) != tt.wantErr {
				_ = tx.Rollback()
				t.Errorf("RetrieveStripeMember() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			_ = tx.Commit()

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}
