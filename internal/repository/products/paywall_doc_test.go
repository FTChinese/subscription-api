package products

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/test"
	"testing"
)

func TestEnv_CreatePaywallDoc(t *testing.T) {

	env := New(db.MockMySQL())

	type args struct {
		pwb reader.PaywallDoc
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Create paywall doc",
			args: args{
				pwb: reader.PaywallDoc{
					ID:          0,
					DailyBanner: test.NewDailyBanner(),
					PromoBanner: reader.BannerJSON{},
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
