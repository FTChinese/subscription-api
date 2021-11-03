package config

import (
	"errors"
	"log"

	"github.com/spf13/viper"
)

// AuthKeys is used to contain api access token set authorization header.
// Those keys are always comes in pair, one for development and one for production.
type AuthKeys struct {
	Dev  string `mapstructure:"dev"`
	Prod string `mapstructure:"prod"`
	name string
}

func (k AuthKeys) Validate() error {
	if k.Dev == "" || k.Prod == "" {
		return errors.New("dev or prod key found")
	}

	return nil
}

func (k AuthKeys) Pick(prod bool) string {
	log.Printf("Using %s for production %t", k.name, prod)

	if prod {
		return k.Prod
	}

	return k.Dev
}

func LoadAuthKeys(name string) (AuthKeys, error) {
	var keys AuthKeys
	err := viper.UnmarshalKey(name, &keys)
	if err != nil {
		return keys, err
	}

	if err := keys.Validate(); err != nil {
		return keys, err
	}

	keys.name = name

	return keys, nil
}

func MustLoadAuthKeys(name string) AuthKeys {
	k, err := LoadAuthKeys(name)
	if err != nil {
		log.Fatalf("cannot get %s: %s", name, err.Error())
	}

	return k
}

func MustLoadStripeAPIKeys() AuthKeys {
	return MustLoadAuthKeys("api_keys.stripe_secret")
}

func MustLoadStripeSigningKey() AuthKeys {
	return MustLoadAuthKeys("api_keys.stripe_webhook_v2")
}

func MustLoadPollingKey() AuthKeys {
	return MustLoadAuthKeys("api_keys.ftc_polling")
}

func MustSubsAPIv1BaseURL() AuthKeys {
	return MustLoadAuthKeys("api_urls.subs_v1")
}
