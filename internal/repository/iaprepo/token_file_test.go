package iaprepo

import (
	"github.com/FTChinese/subscription-api/pkg/apple"
	"github.com/FTChinese/subscription-api/test"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestSaveReceiptToDisk(t *testing.T) {
	type args struct {
		r apple.ReceiptSchema
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Save receipt",
			args: args{
				r: test.MustVerificationResponse().ReceiptToken(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SaveReceiptToDisk(tt.args.r); (err != nil) != tt.wantErr {
				t.Errorf("SaveReceiptToDisk() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadReceiptFromDisk(t *testing.T) {
	rt := test.MustVerificationResponse().ReceiptToken()

	type args struct {
		s apple.BaseSchema
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Load Receipt",
			args: args{
				s: apple.BaseSchema{
					OriginalTransactionID: rt.OriginalTransactionID,
					Environment:           rt.Environment,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadReceiptFromDisk(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadReceiptFromDisk() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.NotNil(t, got)

			t.Logf("Token: %s", got)
		})
	}
}

func TestEnv_SaveReceiptToDB(t *testing.T) {

	env := Env{
		dbs:    test.SplitDB,
		rdb:    test.Redis,
		logger: zaptest.NewLogger(t),
	}

	type args struct {
		r apple.ReceiptSchema
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Save receipt to db",
			args: args{
				r: test.NewPersona().IAPReceiptSchema(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := env.SaveReceiptToDB(tt.args.r); (err != nil) != tt.wantErr {
				t.Errorf("SaveReceiptToDB() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnv_LoadReceiptFromDB(t *testing.T) {
	p := test.NewPersona()
	rt := p.IAPReceiptSchema()
	test.NewRepo().MustSaveIAPReceipt(rt)

	env := Env{
		dbs:    test.SplitDB,
		rdb:    test.Redis,
		logger: zaptest.NewLogger(t),
	}

	type args struct {
		s apple.BaseSchema
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Load receipt from db",
			args: args{
				s: rt.BaseSchema,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := env.LoadReceiptFromDB(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadReceiptFromDB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("%v", got)
		})
	}
}
