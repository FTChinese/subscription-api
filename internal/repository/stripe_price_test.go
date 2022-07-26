package repository

import (
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/test"
	"go.uber.org/zap/zaptest"
	"reflect"
	"testing"
)

func TestStripeRepo_UpsertPrice(t *testing.T) {

	repo := NewStripeRepo(db.MockMySQL(), zaptest.NewLogger(t))

	type args struct {
		p price.StripePrice
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Insert a new price",
			args: args{
				p: price.MockRandomStripePrice(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := repo.UpsertPrice(tt.args.p); (err != nil) != tt.wantErr {
				t.Errorf("UpsertPrice() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStripeRepo_RetrievePrice(t *testing.T) {
	repo := NewStripeRepo(db.MockMySQL(), zaptest.NewLogger(t))
	sp := price.MockRandomStripePrice()

	test.NewRepo().SaveStripePrice(sp)

	type args struct {
		id   string
		live bool
	}
	tests := []struct {
		name    string
		args    args
		want    price.StripePrice
		wantErr bool
	}{
		{
			name: "Retrieve stripe price",
			args: args{
				id:   sp.ID,
				live: false,
			},
			want:    sp,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := repo.RetrievePrice(tt.args.id, tt.args.live)
			if (err != nil) != tt.wantErr {
				t.Errorf("RetrievePrice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RetrievePrice() got = %v, want %v", got, tt.want)
			}
		})
	}
}
