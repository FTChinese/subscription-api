package config

import (
	"bytes"
	"github.com/spf13/viper"
)

// SetupViper config viper
// Deprecated.
func SetupViper() error {
	viper.SetConfigName("api")
	viper.AddConfigPath("$HOME/config")

	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	return nil
}

// MustSetupViper config viper or panic
// Deprecated
func MustSetupViper() {
	if err := SetupViper(); err != nil {
		panic(err)
	}
}

func SetupViperV2(b []byte) error {
	viper.SetConfigType("toml")

	err := viper.ReadConfig(bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	return nil
}

func MustSetupViperV2(b []byte) {
	if err := SetupViperV2(b); err != nil {
		panic(err)
	}
}
