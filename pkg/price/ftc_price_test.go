package price

import (
	"testing"

	"github.com/FTChinese/subscription-api/pkg/conv"
)

func TestFtcPrice_uniqueFeatures(t *testing.T) {

	tests := []struct {
		name   string
		fields FtcPrice
		want   string
	}{
		{
			name:   "standard year intro",
			fields: MockFtcStdIntroPrice,
			want:   "ftc.standard.null.one_time.sandbox",
		},
		{
			name:   "standard year",
			fields: MockFtcStdYearPrice,
			want:   "ftc.standard.year.recurring.sandbox",
		},
		{
			name:   "standard month",
			fields: MockFtcStdMonthPrice,
			want:   "ftc.standard.month.recurring.sandbox",
		},
		{
			name:   "premium year",
			fields: MockFtcPrmPrice,
			want:   "ftc.premium.year.recurring.sandbox",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.fields
			if got := p.uniqueFeatures(); got != tt.want {
				t.Errorf("FtcPrice.uniqueFeatures() = %v, want %v", got, tt.want)
			}

			t.Logf("Active id: %s", p.ActiveID())
		})
	}
}

func TestActivePriceID(t *testing.T) {
	features := []string{
		// abdfb3a810c09e2c2fb9b808f9db6903
		"ftc.standard.null.one_time.sandbox",
		// f7471cbdfef5975e1884e8f28b514352
		"ftc.standard.null.one_time.live",
		// 8923031c43fa7ca5e9a2a223e8c52025
		"ftc.standard.year.recurring.sandbox",
		// a33940fde569add4b9d07c879bc1e839
		"ftc.standard.year.recurring.live",
		// 5e97436472c348b136d22e0477cfe5bc
		"ftc.standard.month.recurring.sandbox",
		// 0db797ffe5f1cee42f76a334b9d53280
		"ftc.standard.month.recurring.live",
		// 5c39c27ccaa2943984431b101ecc069c
		"stripe.standard.null.one_time.sandbox",
		// dea804aada9ceb744f0f1af1a5463f5b
		"stripe.standard.null.one_time.live",
		// 7f0a0e20545c40896a613b6b5395f36c
		"stripe.standard.year.recurring.sandbox",
		// c86d1d93c4e127c5ee41f81a96932f38
		"stripe.standard.year.recurring.live",
		// 5f8eb8a736d27cf81114415871314fe0
		"stripe.standard.month.recurring.sandbox",
		// f5f396bceb5b40eeaadb9f843b0355c0
		"stripe.standard.month.recurring.live",
		// f798fe71178e9b93161113a60ee3667b
		"ftc.premium.year.recurring.sandbox",
		// acfc28992bd2dce1f9bd1e33b99d0995
		"stripe.premium.year.recurring.sandbox",
	}

	for _, v := range features {
		t.Logf("%s: %s", v, conv.NewMD5Sum(v))
	}
}
