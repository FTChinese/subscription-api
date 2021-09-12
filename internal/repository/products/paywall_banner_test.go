package products

import (
	"github.com/FTChinese/subscription-api/test"
	"testing"
)

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
