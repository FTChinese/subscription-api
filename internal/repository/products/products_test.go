package products

import (
	"github.com/FTChinese/subscription-api/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

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

			t.Logf("%+v", got)

			assert.Len(t, got, 2)
		})
	}
}
