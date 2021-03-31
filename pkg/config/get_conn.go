package config

import (
	"github.com/FTChinese/go-rest/connect"
	"github.com/spf13/viper"
)

func GetConn(key string) (connect.Connect, error) {
	var conn connect.Connect
	err := viper.UnmarshalKey(key, &conn)
	if err != nil {
		return connect.Connect{}, err
	}

	return conn, nil
}
