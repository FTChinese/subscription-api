package config

import (
	"github.com/FTChinese/go-rest/connect"
	"log"
)

func mustMySQLConn(key string, prod bool) connect.Connect {
	var conn connect.Connect
	var err error

	if prod {
		conn, err = GetConn(key)
	} else {
		conn, err = GetConn("mysql.dev")
	}

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Using mysql server %s. Production: %t", conn.Host, prod)

	return conn
}

func MustMySQLMasterConn(prod bool) connect.Connect {
	return mustMySQLConn("mysql.master", prod)
}

func MustMySQLAPIConn(prod bool) connect.Connect {
	return mustMySQLConn("mysql.apisuper", prod)
}
