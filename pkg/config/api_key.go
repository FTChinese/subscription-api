package config

import (
	"github.com/spf13/viper"
	"log"
)

// API holds api related access keys or urls.
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
