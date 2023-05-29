package repository

import (
	"reflect"
	"testing"

	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/test"
	"go.uber.org/zap/zaptest"
)

func TestStripeRepo_countPrices(t *testing.T) {
	repo := NewStripeRepo(db.MockMySQL(), zaptest.NewLogger(t))

	type args struct {
		live bool
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "Count stripe prices",
			args: args{
				live: false,
			},
			want:    3,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := repo.countPrices(tt.args.live)
			if (err != nil) != tt.wantErr {
				t.Errorf("StripeRepo.countPrices() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// if got != tt.want {
			// 	t.Errorf("StripeRepo.countPrices() = %v, want %v", got, tt.want)
			// }

			t.Logf("got %d", got)
		})
	}
}

func TestStripeRepo_listPrices(t *testing.T) {
	repo := NewStripeRepo(db.MockMySQL(), zaptest.NewLogger(t))

	type args struct {
		live bool
		p    gorest.Pagination
	}
	tests := []struct {
		name    string
		args    args
		want    []price.StripePrice
		wantErr bool
	}{
		{
			name: "list prices with pagination",
			args: args{
				live: false,
				p:    gorest.NewPagination(1, 10),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := repo.listPrices(tt.args.live, tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("StripeRepo.listPrices() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("got %d items", len(got))
		})
	}
}

func TestStripeRepo_ListPricesPaged(t *testing.T) {
	repo := NewStripeRepo(db.MockMySQL(), zaptest.NewLogger(t))

	type args struct {
		live bool
		p    gorest.Pagination
	}
	tests := []struct {
		name    string
		args    args
		want    pkg.PagedData[price.StripePrice]
		wantErr bool
	}{
		{
			name: "List stripe prices paged",
			args: args{
				live: false,
				p:    gorest.NewPagination(1, 10),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := repo.ListPricesPaged(tt.args.live, tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("StripeRepo.ListPrices() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("StripeRepo.ListPrices() = %v, want %v", got, tt.want)
			// }

			t.Logf("Total %d, current page %d, item on current page %d, got items %d", got.Total, got.Page, got.Limit, len(got.Data))

			t.Logf("%s", faker.MustMarshalIndent(got))
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

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}

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
