package products

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/pw"
	"github.com/FTChinese/subscription-api/test"
	"testing"
)

func TestEnv_CreatePaywallDoc(t *testing.T) {

	env := newTestEnv(db.MockMySQL(), nil)

	type args struct {
		pwb pw.PaywallDoc
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Create paywall doc",
			args: args{
				pwb: pw.PaywallDoc{
					ID:          0,
					DailyBanner: test.NewDailyBanner(),
					PromoBanner: pw.BannerJSON{},
					LiveMode:    false,
					CreatedUTC:  chrono.TimeNow(),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := env.CreatePaywallDoc(tt.args.pwb)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreatePaywallDoc() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("Last insert id %d", got)
		})
	}
}

func TestEnv_RetrievePaywallDoc(t *testing.T) {

	env := newTestEnv(db.MockMySQL(), nil)

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
