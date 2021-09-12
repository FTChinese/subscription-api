package products

import (
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/pw"
	"github.com/FTChinese/subscription-api/test"
	"testing"
)

func TestEnv_retrievePaywall(t *testing.T) {
	env := NewEnv(db.MockMySQL(), nil)

	type args struct {
		live bool
	}
	tests := []struct {
		name    string
		args    args
		want    pw.Paywall
		wantErr bool
	}{
		{
			name: "Retrieve paywall",
			args: args{
				live: true,
			},
			want:    pw.Paywall{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := env.retrievePaywall(tt.args.live)
			if (err != nil) != tt.wantErr {
				t.Errorf("retrievePaywall() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("retrievePaywall() got = %v, want %v", got, tt.want)
			//}
			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}

func TestEnv_retrieveBanner(t *testing.T) {

	env := Env{
		dbs:   test.SplitDB,
		cache: test.Cache,
	}
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "Load banner",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := env.retrieveBanner()
			if (err != nil) != tt.wantErr {
				t.Errorf("retrieveBanner() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("%+v", got)
		})
	}
}

func TestEnv_retrieveActiveProducts(t *testing.T) {
	env := Env{
		dbs:   test.SplitDB,
		cache: test.Cache,
	}

	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name: "Load products for paywall",

			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := env.retrieveActiveProducts()
			if (err != nil) != tt.wantErr {
				t.Errorf("retrieveActiveProducts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("%+s", faker.MustMarshalIndent(got))
		})
	}
}

func TestEnv_listActivePrices(t *testing.T) {

	env := NewEnv(db.MockMySQL(), nil)

	type args struct {
		live bool
	}
	tests := []struct {
		name    string
		args    args
		want    []price.FtcPrice
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
