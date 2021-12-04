package shared

import (
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/pw"
	"testing"
)

func TestEnv_retrievePaywall(t *testing.T) {
	env := PaywallCommon{
		dbs:   db.MockMySQL(),
		cache: nil,
	}

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

func TestEnv_retrieveActiveProducts(t *testing.T) {
	env := PaywallCommon{
		dbs:   db.MockMySQL(),
		cache: nil,
	}

	type args struct {
		live bool
	}

	tests := []struct {
		name    string
		args    args
		want    []pw.Product
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

	env := PaywallCommon{
		dbs:   db.MockMySQL(),
		cache: nil,
	}

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

func TestEnv_LoadPaywall(t *testing.T) {

	env := PaywallCommon{
		dbs:   db.MockMySQL(),
		cache: nil,
	}

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
			name: "Paywall live",
			args: args{
				live: true,
			},
			want:    pw.Paywall{},
			wantErr: false,
		},
		{
			name: "Paywall sandbox",
			args: args{
				live: true,
			},
			want:    pw.Paywall{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := env.LoadPaywall(tt.args.live)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadPaywall() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("LoadPaywall() got = %v, want %v", got, tt.want)
			//}

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}

func TestEnv_RetrievePaywallDoc(t *testing.T) {

	env := NewPaywallCommon(db.MockMySQL(), nil)

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
