package config

import (
	"bytes"
	"github.com/spf13/viper"
)

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
