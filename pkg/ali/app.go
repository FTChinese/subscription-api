package ali

import (
	"errors"
	"github.com/spf13/viper"
)

// App contains the data of an Alipay app.
type App struct {
	ID         string `mapstructure:"app_id"`
	PublicKey  string `mapstructure:"public_key"`
	PrivateKey string `mapstructure:"private_key"`
}

func NewApp(key string) (App, error) {
	var app App

	if err := viper.UnmarshalKey(key, &app); err != nil {
		return App{}, err
	}

	if err := app.Validate(); err != nil {
		return App{}, err
	}

	return app, nil
}

func MustInitApp() App {
	app, err := NewApp("alipay")
	if err != nil {
		panic(err)
	}

	return app
}

func (a App) Validate() error {
	if a.ID == "" || a.PublicKey == "" || a.PrivateKey == "" {
		return errors.New("ali.App has empty fields")
	}

	return nil
}
