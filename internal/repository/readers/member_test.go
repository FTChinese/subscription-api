package readers

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/test"
	"github.com/guregu/null"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestEnv_RetrieveMember(t *testing.T) {
	m1 := reader.NewMockMemberBuilderV2(enum.AccountKindFtc).Build()
	m2 := reader.NewMockMemberBuilderV2(enum.AccountKindWx).Build()
	m3 := reader.NewMockMemberBuilderV2(enum.AccountKindLinked).Build()

	repo := test.NewRepo()
	repo.MustSaveMembership(m1)
	repo.MustSaveMembership(m2)
	repo.MustSaveMembership(m3)

	type fields struct {
		DBs    db.ReadWriteMyDBs
		Logger *zap.Logger
	}
	type args struct {
		compoundID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    reader.Membership
		wantErr bool
	}{
		{
			name: "Retrieve by uuid",
			fields: fields{
				DBs:    test.SplitDB,
				Logger: zaptest.NewLogger(t),
			},
			args: args{
				compoundID: m1.CompoundID,
			},
			wantErr: false,
		},
		{
			name: "Retrieve by uuid",
			fields: fields{
				DBs:    test.SplitDB,
				Logger: zaptest.NewLogger(t),
			},
			args: args{
				compoundID: m2.CompoundID,
			},
			wantErr: false,
		},
		{
			name: "Retrieve linked by any of the ids",
			fields: fields{
				DBs:    test.SplitDB,
				Logger: zaptest.NewLogger(t),
			},
			args: args{
				compoundID: m3.FtcID.String,
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
			got, err := env.RetrieveMember(tt.args.compoundID)
			if (err != nil) != tt.wantErr {
				t.Errorf("RetrieveMember() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("RetrieveMember() got = %v, want %v", got, tt.want)
			//}

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}

func TestEnv_RetrieveAppleMember(t *testing.T) {
	m := reader.NewMockMemberBuilderV2(enum.AccountKindFtc).WithPayMethod(enum.PayMethodApple).Build()

	test.NewRepo().MustSaveMembership(m)

	type fields struct {
		DBs    db.ReadWriteMyDBs
		Logger *zap.Logger
	}
	type args struct {
		txID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    reader.Membership
		wantErr bool
	}{
		{
			name: "Retrieve apple membership",
			fields: fields{
				DBs:    test.SplitDB,
				Logger: zaptest.NewLogger(t),
			},
			args: args{
				txID: m.AppleSubsID.String,
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
			got, err := env.RetrieveAppleMember(tt.args.txID)
			if (err != nil) != tt.wantErr {
				t.Errorf("RetrieveAppleMember() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("RetrieveAppleMember() got = %v, want %v", got, tt.want)
			//}

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}

func TestEnv_ArchiveMember(t *testing.T) {
	m := reader.NewMockMemberBuilderV2(enum.AccountKindFtc).Build()

	type fields struct {
		DBs    db.ReadWriteMyDBs
		Logger *zap.Logger
	}
	type args struct {
		snapshot reader.MemberSnapshot
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Archive membership",
			fields: fields{
				DBs:    test.SplitDB,
				Logger: zaptest.NewLogger(t),
			},
			args: args{
				snapshot: m.Snapshot(reader.Archiver{
					Name:   reader.ArchiveNameOrder,
					Action: reader.ArchiveActionCreate,
				}),
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
			if err := env.ArchiveMember(tt.args.snapshot); (err != nil) != tt.wantErr {
				t.Errorf("ArchiveMember() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnv_VersionMembership(t *testing.T) {

	env := New(db.MockMySQL(), zaptest.NewLogger(t))

	p := test.NewPersona()

	type args struct {
		v reader.MembershipVersioned
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Save membership version",
			args: args{
				v: reader.MembershipVersioned{
					ID: ids.SnapshotID(),
					AnteChange: reader.MembershipJSON{
						Membership: p.MemberBuilder().Build(),
					},
					CreatedBy:        null.StringFrom(reader.NewStripeArchiver(reader.ArchiveActionWebhook).String()),
					CreatedUTC:       chrono.TimeNow(),
					B2BTransactionID: null.String{},
					PostChange: reader.MembershipJSON{
						Membership: p.MemberBuilder().Build(),
					},
					RetailOrderID: null.String{},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := env.VersionMembership(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("VersionMembership() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
