package config

import (
	"github.com/spf13/viper"
	"log"
	"strings"
)

type Address struct {
	Dev  string `mapstructure:"development"`
	Prod string `mapstructure:"production"`
	key  string
}

func MustAddress(key string) Address {
	var a Address
	err := viper.UnmarshalKey(key, &a)
	if err != nil {
		log.Fatal(err)
	}

	a.key = key

	return a
}

func MustKafkaAddress() Address {
	return MustAddress("kafka")
}

func MustRedisAddress() Address {
	return MustAddress("redis")
}

func (a Address) Pick(prod bool) string {
	if prod {
		log.Printf("Using production %s", a.key)
		return a.Prod
	}

	log.Printf("Using development %s", a.key)
	return a.Dev
}

func (a Address) PickSlice(prod bool) []string {
	return strings.Split(a.Pick(prod), ",")
}
