package repository

import (
	"github.com/FTChinese/subscription-api/internal/pkg/stripe"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/test"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestStripeRepo_UpsertPaymentMethod(t *testing.T) {
	repo := NewStripeRepo(db.MockMySQL(), zaptest.NewLogger(t))

	type args struct {
		pm stripe.PaymentMethod
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "",
			args: args{
				pm: stripe.MockPaymentMethod(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := repo.UpsertPaymentMethod(tt.args.pm); (err != nil) != tt.wantErr {
				t.Errorf("UpsertPaymentMethod() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStripeRepo_RetrievePaymentMethod(t *testing.T) {
	repo := NewStripeRepo(db.MockMySQL(), zaptest.NewLogger(t))

	pm := stripe.MockPaymentMethod()

	test.NewRepo().SaveStripePM(pm)

	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		want    stripe.PaymentMethod
		wantErr bool
	}{
		{
			name: "",
			args: args{
				id: pm.ID,
			},
			want:    stripe.PaymentMethod{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := repo.RetrievePaymentMethod(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("RetrievePaymentMethod() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("RetrievePaymentMethod() got = %v, want %v", got, tt.want)
			//}

			t.Logf("%v", got)
		})
	}
}
