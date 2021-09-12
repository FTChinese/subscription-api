package config

import (
	"bytes"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"path/filepath"
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
	//viper.SetConfigName("api")
	//viper.AddConfigPath("$HOME/config")

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

func ReadConfigFile() ([]byte, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	return ioutil.ReadFile(filepath.Join(home, "config", "api.toml"))
}

func MustReadConfigFile() []byte {
	b, err := ReadConfigFile()
	if err != nil {
		panic(err)
	}

	return b
}
