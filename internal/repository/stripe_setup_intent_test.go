package repository

import (
	"github.com/FTChinese/subscription-api/internal/pkg/stripe"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/test"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestStripeRepo_UpsertSetupIntent(t *testing.T) {

	repo := NewStripeRepo(db.MockMySQL(), zaptest.NewLogger(t))

	type args struct {
		si stripe.SetupIntent
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "",
			args: args{
				si: test.StripeSetupIntent(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := repo.UpsertSetupIntent(tt.args.si); (err != nil) != tt.wantErr {
				t.Errorf("UpsertSetupIntent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStripeRepo_RetrieveSetupIntent(t *testing.T) {
	repo := NewStripeRepo(db.MockMySQL(), zaptest.NewLogger(t))

	si := test.StripeSetupIntent()

	test.NewRepo().SaveStripeSetupIntent(si)

	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		want    stripe.SetupIntent
		wantErr bool
	}{
		{
			name: "",
			args: args{
				id: si.ID,
			},
			want:    stripe.SetupIntent{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := repo.RetrieveSetupIntent(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("RetrieveSetupIntent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("RetrieveSetupIntent() got = %v, want %v", got, tt.want)
			//}

			t.Logf("%v", got)
		})
	}
}
