package striperepo

import (
	"encoding/json"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/stripe"
	"github.com/FTChinese/subscription-api/test"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"github.com/jmoiron/sqlx"
	stripeSdk "github.com/stripe/stripe-go/v72"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"testing"
)

func MustNewSubs() stripe.Subs {
	var ss stripeSdk.Subscription
	if err := json.Unmarshal([]byte(faker.StripeSubs), &ss); err != nil {
		panic(err)
	}

	subs, err := stripe.NewSubs(&ss, reader.MemberID{
		CompoundID: "",
		FtcID:      null.StringFrom(uuid.New().String()),
		UnionID:    null.String{},
	}.MustNormalize())

	if err != nil {
		panic(err)
	}

	return subs
}

func TestEnv_UpsertSubs(t *testing.T) {

	type fields struct {
		db     *sqlx.DB
		client Client
		logger *zap.Logger
	}
	type args struct {
		s stripe.Subs
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Save subs",
			fields: fields{
				db:     test.DB,
				logger: zaptest.NewLogger(t),
			},
			args: args{
				s: MustNewSubs(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				db:     tt.fields.db,
				logger: tt.fields.logger,
			}
			if err := env.UpsertSubs(tt.args.s); (err != nil) != tt.wantErr {
				t.Errorf("UpsertSubs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnv_RetrieveSubs(t *testing.T) {
	type fields struct {
		db     *sqlx.DB
		client Client
		logger *zap.Logger
	}
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    stripe.Subs
		wantErr bool
	}{
		{
			name: "Retrieve subs",
			fields: fields{
				db:     test.DB,
				logger: zaptest.NewLogger(t),
			},
			args: args{
				id: "sub_IX3JAkik1JKDzW",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				db:     tt.fields.db,
				client: tt.fields.client,
				logger: tt.fields.logger,
			}
			got, err := env.RetrieveSubs(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSubs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}

func TestEnv_SubsExists(t *testing.T) {
	type fields struct {
		db     *sqlx.DB
		client Client
		logger *zap.Logger
	}
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "Subs exists",
			fields: fields{
				db:     test.DB,
				logger: zaptest.NewLogger(t),
			},
			args: args{
				id: "sub_IX3JAkik1JKDzW",
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				db:     tt.fields.db,
				client: tt.fields.client,
				logger: tt.fields.logger,
			}
			got, err := env.SubsExists(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("SubsExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SubsExists() got = %v, want %v", got, tt.want)
			}
		})
	}
}
