package striperepo

import (
	"encoding/json"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/internal/repository/readers"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/stripe"
	"github.com/FTChinese/subscription-api/test"
	"github.com/google/uuid"
	"github.com/guregu/null"
	stripeSdk "github.com/stripe/stripe-go/v72"
	"go.uber.org/zap/zaptest"
	"testing"
)

func MustNewSubs() stripe.Subs {
	var ss stripeSdk.Subscription
	if err := json.Unmarshal([]byte(faker.StripeSubs), &ss); err != nil {
		panic(err)
	}

	subs, err := stripe.NewSubs(&ss, ids.UserIDs{
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

	env := Env{
		Env:    readers.New(test.SplitDB, zaptest.NewLogger(t)),
		client: Client{},
		logger: zaptest.NewLogger(t),
	}

	type args struct {
		s stripe.Subs
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Save subs",
			args: args{
				s: MustNewSubs(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := env.UpsertSubs(tt.args.s); (err != nil) != tt.wantErr {
				t.Errorf("UpsertSubs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnv_RetrieveSubs(t *testing.T) {
	env := Env{
		Env:    readers.New(test.SplitDB, zaptest.NewLogger(t)),
		client: Client{},
		logger: zaptest.NewLogger(t),
	}

	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		want    stripe.Subs
		wantErr bool
	}{
		{
			name: "Retrieve subs",
			args: args{
				id: "sub_IX3JAkik1JKDzW",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
	env := Env{
		Env:    readers.New(test.SplitDB, zaptest.NewLogger(t)),
		client: Client{},
		logger: zaptest.NewLogger(t),
	}

	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "Subs exists",
			args: args{
				id: "sub_IX3JAkik1JKDzW",
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

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
