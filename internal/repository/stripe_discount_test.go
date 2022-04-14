package repository

import (
	"github.com/FTChinese/subscription-api/internal/pkg/stripe"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/test"
	"go.uber.org/zap/zaptest"
	"reflect"
	"testing"
)

func TestStripeRepo_UpsertDiscount(t *testing.T) {
	repo := NewStripeRepo(db.MockMySQL(), zaptest.NewLogger(t))

	d := stripe.MockRandomDiscount()

	type args struct {
		d stripe.Discount
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "",
			args: args{
				d: d,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := repo.UpsertDiscount(tt.args.d); (err != nil) != tt.wantErr {
				t.Errorf("UpsertDiscount() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStripeRepo_RetrieveDiscount(t *testing.T) {
	repo := NewStripeRepo(db.MockMySQL(), zaptest.NewLogger(t))

	d := stripe.MockRandomDiscount()

	test.NewRepo().SaveStripeDiscount(d)

	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		want    stripe.Discount
		wantErr bool
	}{
		{
			name: "",
			args: args{
				id: d.ID,
			},
			want:    d,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := repo.RetrieveDiscount(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("RetrieveDiscount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RetrieveDiscount() got = %v, want %v", got, tt.want)
			}
		})
	}
}
