package cache

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *redis.Client
	ctx    context.Context
}

func New() *Cache {
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	redisPassword := getEnv("REDIS_PASSWORD", "")
	redisDB := 0 

	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})

	ctx := context.Background()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		println("Warning: Redis connection failed:", err.Error())
	}

	return &Cache{
		client: client,
		ctx:    ctx,
	}
}

func (c *Cache) Set(key string, value interface{}, expiration time.Duration) {
	jsonData, err := json.Marshal(value)
	if err != nil {
		println("Cache serialization error:", err.Error())
		return
	}

	err = c.client.Set(c.ctx, key, jsonData, expiration).Err()
	if err != nil {
		println("Cache set error:", err.Error())
	}
}

func (c *Cache) Get(key string) (interface{}, bool) {
	jsonData, err := c.client.Get(c.ctx, key).Bytes()
	if err != nil {
		if err != redis.Nil {
			println("Cache get error:", err.Error())
		}
		return nil, false
	}

	var value interface{}
	err = json.Unmarshal(jsonData, &value)
	if err != nil {
		println("Cache deserialization error:", err.Error())
		return nil, false
	}

	return value, true
}

func (c *Cache) Delete(key string) {
	err := c.client.Del(c.ctx, key).Err()
	if err != nil {
		println("Cache delete error:", err.Error())
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
