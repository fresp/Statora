package database

import (
	"context"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

var redisClient *redis.Client

func ConnectRedis(addr string) error {
	var opt *redis.Options
	var err error

	// ⬇️ Support redis:// URI
	if len(addr) > 8 && addr[:8] == "redis://" {
		opt, err = redis.ParseURL(addr)
		if err != nil {
			return err
		}
	} else {
		// fallback lama (host:port)
		opt = &redis.Options{
			Addr: addr,
			DB:   0,
		}
	}

	// ⬇️ Optional tuning (minimal)
	opt.PoolSize = 20
	opt.MinIdleConns = 5

	rdb := redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

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