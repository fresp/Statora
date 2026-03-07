package database

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
)

var redisClient *redis.Client

func ConnectRedis(addr string) error {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   0,
	})

	ctx := context.Background()
	if _, err := rdb.Ping(ctx).Result(); err != nil {
		return err
	}

	redisClient = rdb
	log.Printf("Connected to Redis: %s", addr)
	return nil
}

func GetRedis() *redis.Client {
	return redisClient
}
