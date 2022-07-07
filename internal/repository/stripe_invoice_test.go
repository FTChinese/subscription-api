package repository

import (
	"github.com/FTChinese/subscription-api/internal/pkg/stripe"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/test"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestStripeRepo_UpsertInvoice(t *testing.T) {
	repo := NewStripeRepo(db.MockMySQL(), zaptest.NewLogger(t))

	type args struct {
		i stripe.Invoice
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Upsert invoice",
			args: args{
				i: test.StripeInvoice(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := repo.UpsertInvoice(tt.args.i); (err != nil) != tt.wantErr {
				t.Errorf("UpsertInvoice() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStripeRepo_RetrieveInvoice(t *testing.T) {
	repo := NewStripeRepo(db.MockMySQL(), zaptest.NewLogger(t))

	inv := test.StripeInvoice()

	_ = repo.UpsertInvoice(inv)

	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		want    stripe.Invoice
		wantErr bool
	}{
		{
			name: "Retrieve invoice",
			args: args{
				id: inv.ID,
			},
			want:    stripe.Invoice{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := repo.RetrieveInvoice(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("RetrieveInvoice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("RetrieveInvoice() got = %v, want %v", got, tt.want)
			//}

			t.Logf("%v", got)
		})
	}
}
