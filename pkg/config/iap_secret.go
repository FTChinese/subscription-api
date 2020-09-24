package config

import "github.com/spf13/viper"

func MustIAPSecret() string {
	pw := viper.GetString("apple.receipt_password")
	if pw == "" {
		panic("empty receipt verification password")
	}

	return pw
}
