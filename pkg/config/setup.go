package config

import "github.com/spf13/viper"

func SetupViper() error {
	viper.SetConfigName("api")
	viper.AddConfigPath("$HOME/config")

	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	return nil
}

func MustSetupViper() {
	if err := SetupViper(); err != nil {
		panic(err)
	}
}
