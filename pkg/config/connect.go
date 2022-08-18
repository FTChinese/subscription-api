package config

import (
	"github.com/spf13/viper"
	"log"
)

// Connect represents a connection to a server or database.
type Connect struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	User string `mapstructure:"user"`
	Pass string `mapstructure:"pass"`
}

func GetConn(key string) (Connect, error) {
	var conn Connect
	err := viper.UnmarshalKey(key, &conn)
	if err != nil {
		return Connect{}, err
	}

	return conn, nil
}

func MustMySQLConn(key string) Connect {
	conn, err := GetConn(key)

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Got MySQL %s ip %s", key, conn.Host)

	return conn
}

func MustMySQLReadConn() Connect {
	return MustMySQLConn("mysql.read")
}

func MustMySQLWriteConn() Connect {
	return MustMySQLConn("mysql.write")
}

func MustMySQLDeleteConn() Connect {
	return MustMySQLConn("mysql.delete")
}

func MustGetHanqiConn() Connect {
	var conn Connect
	err := viper.UnmarshalKey("email.hanqi", &conn)
	if err != nil {
		log.Fatal(err)
	}

	return conn
}
