package iaprepo

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/apple"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/test"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"testing"
	"time"
)

func TestEnv_SaveSubs(t *testing.T) {
	p := test.NewPersona().SetPayMethod(enum.PayMethodApple)

	env := NewEnv(test.DB, test.Redis, zaptest.NewLogger(t))

	p2 := test.NewPersona().
		SetPayMethod(enum.PayMethodApple)

	m := p2.Membership()
	m.ExpireDate = chrono.DateFrom(time.Now().AddDate(0, -6, 0))
	test.NewRepo().MustSaveMembership(m)

	type args struct {
		s apple.Subscription
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Save unlinked iap subscription",
			args: args{
				s: p.IAPSubs(),
			},
			wantErr: false,
		},
		{
			name: "Save linked iap subscription",
			args: args{
				s: p2.IAPSubs(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := env.SaveSubs(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("SaveSubs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("%v", got)
		})
	}
}

func TestEnv_updateMembership(t *testing.T) {
	userID := uuid.New().String()
	txID := faker.GenAppleSubID()

	current := reader.NewMockMemberBuilder(userID).
		WithPayMethod(enum.PayMethodApple).
		WithIapID(txID).
		Build()
	test.NewRepo().MustSaveMembership(current)

	type fields struct {
		db     *sqlx.DB
		logger *zap.Logger
	}
	type args struct {
		s apple.Subscription
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    apple.SubsResult
		wantErr bool
	}{
		{
			name: "Update membership",
			fields: fields{
				db:     test.DB,
				logger: zaptest.NewLogger(t),
			},
			args: args{
				s: apple.NewMockSubsBuilder(userID).
					WithOriginalTxID(txID).
					Build(),
			},
			want:    apple.SubsResult{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				db:     tt.fields.db,
				logger: tt.fields.logger,
			}
			got, err := env.updateMembership(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("updateMembership() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("updateMembership() got = %v, want %v", got, tt.want)
			//}

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}

func TestEnv_LoadSubs(t *testing.T) {
	p := test.NewPersona()
	test.NewRepo().MustSaveIAPSubs(p.IAPSubs())

	type fields struct {
		db *sqlx.DB
	}
	type args struct {
		originalID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Load a subscription",
			fields: fields{
				db: test.DB,
			},
			args:    args{originalID: p.AppleSubID},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				db: tt.fields.db,
			}
			got, err := env.LoadSubs(tt.args.originalID)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			t.Logf("%s", got.Environment)
		})
	}
}

func TestEnv_countSubs(t *testing.T) {
	p := test.NewPersona()
	test.NewRepo().MustSaveIAPSubs(p.IAPSubsLinked())

	type fields struct {
		db *sqlx.DB
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Count iap subscription",
			fields: fields{
				db: test.DB,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				db: tt.fields.db,
			}
			got, err := env.countSubs(p.FtcID)
			if (err != nil) != tt.wantErr {
				t.Errorf("countSubs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			t.Logf("Total rows %d", got)
		})
	}
}

func TestEnv_listSubs(t *testing.T) {
	p := test.NewPersona()
	test.NewRepo().MustSaveIAPSubs(p.IAPSubsLinked())

	type fields struct {
		db *sqlx.DB
	}
	type args struct {
		ftcID string
		p     gorest.Pagination
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "List subs",
			fields: fields{
				db: test.DB,
			},
			args: args{
				ftcID: p.FtcID,
				p:     gorest.NewPagination(1, 20),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				db: tt.fields.db,
			}
			got, err := env.listSubs(tt.args.ftcID, tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("listSubs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}

func TestEnv_ListSubs(t *testing.T) {
	p := test.NewPersona()
	test.NewRepo().MustSaveIAPSubs(p.IAPSubsLinked())

	env := NewEnv(test.DB, test.Redis, zaptest.NewLogger(t))

	t.Logf("Create IAP %s", p.AppleSubID)

	type args struct {
		ftcID string
		p     gorest.Pagination
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Count and list subs",
			args: args{
				ftcID: p.FtcID,
				p:     gorest.NewPagination(1, 20),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := env.ListSubs(tt.args.ftcID, tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListSubs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.NoError(t, err)

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}
