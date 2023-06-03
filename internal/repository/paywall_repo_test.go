package repository

import (
	"testing"

	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/reader"
)

func TestEnv_retrievePaywall(t *testing.T) {
	env := PaywallRepo{
		dbs: db.MockMySQL(),
	}

	type args struct {
		live bool
	}
	tests := []struct {
		name    string
		args    args
		want    reader.Paywall
		wantErr bool
	}{
		{
			name: "Retrieve paywall",
			args: args{
				live: true,
			},
			want:    reader.Paywall{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := env.RetrievePaywall(tt.args.live)
			if (err != nil) != tt.wantErr {
				t.Errorf("RetrievePaywall() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("RetrievePaywall() got = %v, want %v", got, tt.want)
			//}
			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}

func TestEnv_retrieveActiveProducts(t *testing.T) {
	env := PaywallRepo{
		dbs: db.MockMySQL(),
	}

	type args struct {
		live bool
	}

	tests := []struct {
		name    string
		args    args
		want    []reader.Product
		wantErr bool
	}{
		{
			name: "Load active product for sandbox",
			args: args{
				live: false,
			},
			wantErr: false,
		},
		{
			name: "Load active product for live",
			args: args{
				live: true,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := env.retrieveActiveProducts(tt.args.live)
			if (err != nil) != tt.wantErr {
				t.Errorf("retrieveActiveProducts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("%+s", faker.MustMarshalIndent(got))
		})
	}
}

func TestEnv_listActivePrices(t *testing.T) {

	env := PaywallRepo{
		dbs: db.MockMySQL(),
	}

	type args struct {
		live bool
	}
	tests := []struct {
		name    string
		args    args
		want    []reader.PaywallPrice
		wantErr bool
	}{
		{
			name: "Live mode active prices",
			args: args{
				live: true,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "Test mode active prices",
			args: args{
				live: false,
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := env.ListActivePrices(tt.args.live)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListActivePrices() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("ListActivePrices() got = %v, want %v", got, tt.want)
			//}

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}

func TestEnv_RetrievePaywallDoc(t *testing.T) {

	env := NewPaywallRepo(db.MockMySQL())

	type args struct {
		live bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Retrieve paywall doc",
			args: args{
				live: false,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := env.RetrievePaywallDoc(tt.args.live)
			if (err != nil) != tt.wantErr {
				t.Errorf("RetrievePaywallDoc() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}

func TestPaywallRepo_RetrievePaywallV2(t *testing.T) {
	repo := NewPaywallRepo(db.MockMySQL())

	type args struct {
		live bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Retrieve paywall",
			args: args{
				live: false,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.RetrievePaywallV2(tt.args.live)
			if (err != nil) != tt.wantErr {
				t.Errorf("PaywallRepo.RetrievePaywallV2() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}
