package config

import "github.com/spf13/viper"

type Credentials struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

func MustSMSCredentials() Credentials {
	var c Credentials
	err := viper.UnmarshalKey("sms", &c)
	if err != nil {
		panic(err)
	}

	return c
}
