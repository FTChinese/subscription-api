package db

import (
	"github.com/FTChinese/go-rest/connect"
	"github.com/FTChinese/subscription-api/pkg/config"
	"log"
)

func mustMySQLConn(key string, prod bool) connect.Connect {
	var conn connect.Connect
	var err error

	if prod {
		conn, err = config.GetConn(key)
	} else {
		conn, err = config.GetConn("mysql.dev")
	}

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Using mysql server %s. Production: %t", conn.Host, prod)

	return conn
}

func MustMySQLReadConn(prod bool) connect.Connect {
	return mustMySQLConn("mysql.read", prod)
}

func MustMySQLWriteConn(prod bool) connect.Connect {
	return mustMySQLConn("mysql.write", prod)
}

func MustMySQLDeleteConn(prod bool) connect.Connect {
	return mustMySQLConn("mysql.delete", prod)
}
