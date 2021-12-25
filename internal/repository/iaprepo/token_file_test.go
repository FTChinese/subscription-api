package iaprepo

import (
	"github.com/FTChinese/subscription-api/internal/pkg/apple"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/test"
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

			t.Logf("Token: %s", got)
		})
	}
}

func TestEnv_SaveReceiptToDB(t *testing.T) {

	env := New(db.MockMySQL(), nil, zaptest.NewLogger(t))

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
				r: test.NewPersona().IAPBuilder().ReceiptSchema(),
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
	rt := p.IAPBuilder().ReceiptSchema()
	test.NewRepo().MustSaveIAPReceipt(rt)

	env := New(db.MockMySQL(), nil, zaptest.NewLogger(t))

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
