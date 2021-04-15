package readers

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/test"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestEnv_RetrieveMember(t *testing.T) {
	m1 := reader.NewMockMemberBuilderV2(enum.AccountKindFtc).Build()
	m2 := reader.NewMockMemberBuilderV2(enum.AccountKindWx).Build()

	repo := test.NewRepo()
	repo.MustSaveMembership(m1)
	repo.MustSaveMembership(m2)

	type fields struct {
		DBs    db.ReadWriteSplit
		Logger *zap.Logger
	}
	type args struct {
		id pkg.UserIDs
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
				id: m1.UserIDs,
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
				id: m2.UserIDs,
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
			got, err := env.RetrieveMember(tt.args.id)
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
		DBs    db.ReadWriteSplit
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
		DBs    db.ReadWriteSplit
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
					Name:   reader.NameOrder,
					Action: reader.ActionCreate,
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
