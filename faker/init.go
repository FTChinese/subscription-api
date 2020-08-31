// +build !production

package faker

import (
	"github.com/spf13/viper"
	"log"
)

func init() {
	viper.SetConfigName("api")
	viper.AddConfigPath("$HOME/config")

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal(err)
	}
}
