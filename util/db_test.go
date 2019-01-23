package util

import (
	"github.com/spf13/viper"
	"testing"
)

func TestViporConfig(t *testing.T) {
	viper.SetConfigName("api")
	viper.AddConfigPath("$HOME/config")
	err := viper.ReadInConfig()
	if err != nil {
		t.Error(err)
		return
	}

	var conn Conn
	err = viper.UnmarshalKey("mysql.dev", &conn)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("%+v", conn)

	err = viper.UnmarshalKey("mysql.master", &conn)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("%+v", conn)
}