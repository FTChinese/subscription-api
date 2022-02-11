package repository

import (
	"github.com/FTChinese/subscription-api/internal/pkg/stripe"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/test"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestStripeRepo_SaveCustomer(t *testing.T) {
	repo := NewStripeRepo(db.MockMySQL(), zaptest.NewLogger(t))

	type args struct {
		c stripe.Customer
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Save customer",
			args: args{
				c: test.NewPersona().StripeCustomer(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := repo.InsertCustomer(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("InsertCustomer() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			t.Logf("%v", tt.args.c)
		})
	}
}

func TestStripeRepo_UpdateCustomer(t *testing.T) {
	repo := NewStripeRepo(db.MockMySQL(), zaptest.NewLogger(t))

	c := test.NewPersona().StripeCustomer()

	test.NewRepo().SaveStripeCustomer(c)

	type args struct {
		c stripe.Customer
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Update customer",
			args: args{
				c: c,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := repo.UpdateCustomer(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("UpdateCustomer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStripeRepo_RetrieveCustomer(t *testing.T) {
	repo := NewStripeRepo(db.MockMySQL(), zaptest.NewLogger(t))

	c := test.NewPersona().StripeCustomer()

	test.NewRepo().SaveStripeCustomer(c)

	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		want    stripe.Customer
		wantErr bool
	}{
		{
			name: "Retrieve customer",
			args: args{
				id: c.ID,
			},
			want:    stripe.Customer{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := repo.RetrieveCustomer(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("FetchCustomer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("FetchCustomer() got = %v, want %v", got, tt.want)
			//}

			t.Logf("%v", got)
		})
	}
}
