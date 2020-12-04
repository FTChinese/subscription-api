package stripe

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stripe/stripe-go"
	"log"
	"testing"
)

func mustConfigViper() config.BuildConfig {
	viper.SetConfigName("api")
	viper.AddConfigPath("$HOME/config")

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal(err)
	}

	cfg := config.NewBuildConfig(false, false)
	stripe.Key = cfg.MustStripeAPIKey()

	return cfg
}

func Test_newPlanStore(t *testing.T) {
	store := newPlanStore()

	assert.Len(t, store.plans, 6)
	assert.Len(t, store.indexEdition, 6)
	assert.Len(t, store.indexID, 6)
}

func Test_planStore_findByEdition(t *testing.T) {
	p, err := PlanStore.FindByEdition("standard_year", false)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, p.Tier, enum.TierStandard)
}
