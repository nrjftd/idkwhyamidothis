package database

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func InitRedis() {
	// RedisClient = redis.NewClient(&redis.Options{
	// 	Addr:     "localhost: 6379",
	// 	Password: "",
	// 	DB:       0,
	// })
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		host := os.Getenv("REDIS_HOST")
		port := os.Getenv("REDIS_PORT")
		if host == "" {
			host = "redis"
		}
		if port == "" {
			port = "6379"
		}
		redisURL = fmt.Sprintf("%s:%s", host, port)
		log.Println("REDIS_URL is empty, using fallback", redisURL)
	}
	log.Println("REDIS_URL:", redisURL)
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("Failed to parse Redis URL: %v", err)
	}
	RedisClient = redis.NewClient(opt)
	_, err = RedisClient.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Failed to connect Redis: %v", err)

	}
	log.Println("redis connected successfully")
}
