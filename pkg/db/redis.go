package db

import (
	"github.com/go-redis/redis/v8"
)

func NewRedis(addr string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})
}
