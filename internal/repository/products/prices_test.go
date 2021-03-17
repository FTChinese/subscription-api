package products

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEnv_retrieveProductPrices(t *testing.T) {
	env := Env{
		dbs:   test.SplitDB,
		cache: test.Cache,
	}

	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "List paywall plans",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := env.retrieveProductPrices()
			if (err != nil) != tt.wantErr {
				t.Errorf("retrieveProductPrices() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("%+v", got)

			assert.Len(t, got, 3)
		})
	}
}

func TestEnv_RetrievePrice(t *testing.T) {

	env := Env{
		dbs:   test.SplitDB,
		cache: test.Cache,
	}

	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "Load a single plan",
			args:    args{id: "plan_ICMPPM0UXcpZ"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := env.RetrievePrice(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("retrievePlanByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.NotEmpty(t, got.ID)
		})
	}
}

func TestEnv_PlanByEdition(t *testing.T) {
	env := Env{
		dbs:   test.SplitDB,
		cache: test.Cache,
	}

	type args struct {
		e price.Edition
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Load a plan by tier and cycle",
			args: args{
				e: price.Edition{
					Tier:  enum.TierStandard,
					Cycle: enum.CycleYear,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := env.ActivePriceOfEdition(tt.args.e)
			if (err != nil) != tt.wantErr {
				t.Errorf("ActivePriceOfEdition() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.NotEmpty(t, got.ID)
		})
	}
}
