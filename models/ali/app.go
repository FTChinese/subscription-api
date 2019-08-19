package ali

import "errors"

// App contains the data of an Alipay app.
type App struct {
	ID         string `mapstructure:"app_id"`
	PublicKey  string `mapstructure:"public_key"`
	PrivateKey string `mapstructure:"private_key"`
}

func (a App) Ensure() error {
	if a.ID == "" || a.PublicKey == "" || a.PrivateKey == "" {
		return errors.New("ali.App has empty fields")
	}

	return nil
}
