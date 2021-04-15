package txrepo

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/apple"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/test"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"testing"
)

func TestIAPTx_RetrieveAppleMember(t *testing.T) {

	m := reader.NewMockMemberBuilderV2(enum.AccountKindFtc).
		WithPayMethod(enum.PayMethodApple).
		WithIapID(faker.GenAppleSubID()).
		Build()

	repo := test.NewRepo()
	repo.MustSaveMembership(m)

	type fields struct {
		Tx *sqlx.Tx
	}
	type args struct {
		transactionID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Retrieve an IAP member",
			fields: fields{
				Tx: test.SplitDB.Read.MustBegin(),
			},
			args: args{
				transactionID: m.AppleSubsID.String,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := NewIAPTx(tt.fields.Tx)

			got, err := tx.RetrieveAppleMember(tt.args.transactionID)
			if (err != nil) != tt.wantErr {
				t.Errorf("RetrieveAppleMember() error = %v, wantErr %v", err, tt.wantErr)
				_ = tx.Rollback()
				return
			}

			t.Logf("%s", faker.MustMarshalIndent(got))
			_ = tx.Commit()
		})
	}
}

func TestSharedTx_RetrieveAppleSubs(t *testing.T) {

	s := apple.NewMockSubsBuilder("").Build()

	test.NewRepo().MustSaveIAPSubs(s)

	type fields struct {
		Tx *sqlx.Tx
	}
	type args struct {
		origTxID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    apple.Subscription
		wantErr bool
	}{
		{
			name: "Retrieve apple subs",
			fields: fields{
				Tx: test.SplitDB.Write.MustBegin(),
			},
			args: args{
				origTxID: s.OriginalTransactionID,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := SharedTx{
				Tx: tt.fields.Tx,
			}
			got, err := tx.RetrieveAppleSubs(tt.args.origTxID)
			if (err != nil) != tt.wantErr {
				t.Errorf("RetrieveAppleSubs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("RetrieveAppleSubs() got = %v, want %v", got, tt.want)
			//}

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}

func TestSharedTx_LinkAppleSubs(t *testing.T) {
	s := apple.NewMockSubsBuilder("").Build()

	test.NewRepo().MustSaveIAPSubs(s)

	type fields struct {
		Tx *sqlx.Tx
	}
	type args struct {
		link apple.LinkInput
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Set ftc id to apple subscription",
			fields: fields{
				Tx: test.SplitDB.Write.MustBegin(),
			},
			args: args{
				link: apple.LinkInput{
					FtcID:        uuid.New().String(),
					OriginalTxID: s.OriginalTransactionID,
					Force:        false,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := SharedTx{
				Tx: tt.fields.Tx,
			}
			t.Logf("Link %v", tt.args.link)

			if err := tx.LinkAppleSubs(tt.args.link); (err != nil) != tt.wantErr {
				t.Errorf("LinkAppleSubs() error = %v, wantErr %v", err, tt.wantErr)
				_ = tx.Rollback()
				return
			}

			_ = tx.Commit()
		})
	}
}

func TestSharedTx_UnlinkAppleSubs(t *testing.T) {
	s := apple.NewMockSubsBuilder(uuid.New().String()).Build()

	test.NewRepo().MustSaveIAPSubs(s)

	type fields struct {
		Tx *sqlx.Tx
	}
	type args struct {
		link apple.LinkInput
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Unlink ftc id from apple subscription",
			fields: fields{
				Tx: test.SplitDB.Write.MustBegin(),
			},
			args: args{
				link: apple.LinkInput{
					FtcID:        s.FtcUserID.String,
					OriginalTxID: s.OriginalTransactionID,
					Force:        false,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := SharedTx{
				Tx: tt.fields.Tx,
			}
			if err := tx.UnlinkAppleSubs(tt.args.link); (err != nil) != tt.wantErr {
				t.Errorf("UnlinkAppleSubs() error = %v, wantErr %v", err, tt.wantErr)
				_ = tx.Rollback()
				return
			}

			_ = tx.Commit()
		})
	}
}
