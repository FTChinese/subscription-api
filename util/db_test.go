package util

import (
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
	"testing"
)

func init() {
	viper.SetConfigName("api")
	viper.AddConfigPath("$HOME/config")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}

func TestViporConfig(t *testing.T) {
	var conn Conn
	err := viper.UnmarshalKey("mysql.dev", &conn)
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

func TestDBConn(t *testing.T) {
	var c Conn
	err := viper.UnmarshalKey("mysql.dev", &c)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("Connection: %+v", c)

	cfg := &mysql.Config{
		User:   c.User,
		Passwd: c.Pass,
		Net:    "tcp",
		Addr:   fmt.Sprintf("%s:%d", c.Host, c.Port),
		Params: map[string]string{
			"time_zone": `'+00:00'`,
		},
		AllowNativePasswords: true,
	}

	t.Logf("%s", cfg.FormatDSN())
}
