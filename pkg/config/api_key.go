package config

import (
	"errors"
	"github.com/spf13/viper"
	"log"
)

// API holds api related access keys or urls.
// Deprecated.
type API struct {
	Dev  string `mapstructure:"api_key_dev"`
	Prod string `mapstructure:"api_key_prod"`
	name string
}

func MustAPIKey() API {
	var key API

	err := viper.UnmarshalKey("service.iap_polling", &key)
	if err != nil {
		log.Fatal(err)
	}

	key.name = "API key"
	return key
}

func MustAPIBaseURL() API {
	prodURL := viper.GetString("api_url.subscription_v1")

	return API{
		Dev:  "http://localhost:8200",
		Prod: prodURL,
		name: "API base url",
	}
}

func (k API) Pick(prod bool) string {
	if prod {
		log.Printf("Using production %s %s", k.name, k.Prod)
		return k.Prod
	}

	log.Printf("Using development %s %s", k.name, k.Dev)
	return k.Dev
}

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
