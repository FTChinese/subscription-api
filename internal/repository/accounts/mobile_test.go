package accounts

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/ztsms"
	"github.com/FTChinese/subscription-api/test"
	"github.com/brianvoe/gofakeit/v5"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"go.uber.org/zap/zaptest"
	"reflect"
	"testing"
	"time"
)

func TestEnv_SaveSMSVerifier(t *testing.T) {
	env := New(test.SplitDB, zaptest.NewLogger(t))

	faker.SeedGoFake()

	v := ztsms.NewVerifier(gofakeit.Phone(), null.StringFrom(uuid.New().String()))

	type args struct {
		v ztsms.Verifier
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Save mobile verifier",
			args: args{
				v: v,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := env.SaveSMSVerifier(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("SaveSMSVerifier() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnv_RetrieveSMSVerifier(t *testing.T) {
	env := New(test.SplitDB, zaptest.NewLogger(t))

	faker.SeedGoFake()

	v := ztsms.NewVerifier(gofakeit.Phone(), null.StringFrom(uuid.New().String()))

	v.CreatedUTC = chrono.TimeFrom(time.Now().In(time.UTC).Truncate(time.Second))

	_ = env.SaveSMSVerifier(v)

	type args struct {
		params ztsms.VerifierParams
	}
	tests := []struct {
		name    string
		args    args
		want    ztsms.Verifier
		wantErr bool
	}{
		{
			name: "Retrieve sms verifier",
			args: args{
				params: ztsms.VerifierParams{
					Mobile: v.Mobile,
					Code:   v.Code,
				},
			},
			want:    v,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := env.RetrieveSMSVerifier(tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("RetrieveSMSVerifier() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RetrieveSMSVerifier() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnv_SMSVerifierUsed(t *testing.T) {
	env := New(test.SplitDB, zaptest.NewLogger(t))

	v := ztsms.NewVerifier(gofakeit.Phone(), null.StringFrom(uuid.New().String()))

	_ = env.SaveSMSVerifier(v)

	type args struct {
		v ztsms.Verifier
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Flag sms code as used",
			args: args{
				v: v,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			tt.args.v = tt.args.v.WithUsed()

			if err := env.SMSVerifierUsed(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("SMSVerifierUsed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnv_SetMobile(t *testing.T) {

	env := New(test.SplitDB, zaptest.NewLogger(t))
	repo := test.NewRepo()

	type requisite struct {
		kind test.MobileLinkAccountKind
	}
	type args struct {
		params ztsms.MobileUpdater
	}
	tests := []struct {
		name      string
		requisite requisite
		wantErr   bool
	}{
		{
			name: "Row does not exist",
			requisite: requisite{
				kind: test.MobileLinkNoProfile,
			},
			wantErr: false,
		},
		{
			name: "Profile exists with empty mobile",
			requisite: requisite{
				kind: test.MobileLinkHasProfileNoPhone,
			},
			wantErr: false,
		},
		{
			name: "Profile exists with mobile taken",
			requisite: requisite{
				kind: test.MobileLinkHasProfilePhoneSet,
			},
			wantErr: true,
		},
		{
			name: "Profile exists with mobile set",
			requisite: requisite{
				kind: test.MobileLinkHasProfilePhoneTaken,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ftcID, linkParams := repo.GenerateMobileLinkParams(tt.requisite.kind)
			arg := ztsms.MobileUpdater{
				FtcID:  ftcID,
				Mobile: null.StringFrom(linkParams.Mobile),
			}

			if err := env.SetMobile(arg); (err != nil) != tt.wantErr {
				t.Errorf("SetMobile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
