package cache

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *redis.Client
	ctx    context.Context
	items  map[string]cacheItem
	mu     sync.RWMutex
}

type cacheItem struct {
	value      interface{}
	expiration time.Time
}

func New() *Cache {
	// Логи Redis-кэша также сохраняются в тот же файл, что и логи API Gateway
	logFile, err := os.OpenFile("redis_cache.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("Ошибка при открытии файла логов в cache.go: %v", err)
	} else if log.Writer() != logFile {
		// Проверяем, не настроен ли уже вывод в файл
		log.SetOutput(logFile)
		log.Printf("Логирование кэша настроено, логи сохраняются в redis_cache.log")
	}

	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	redisPassword := getEnv("REDIS_PASSWORD", "")
	redisDB := 0

	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})

	ctx := context.Background()

	_, err = client.Ping(ctx).Result()
	if err != nil {
		log.Printf("Warning: Redis connection failed: %s", err.Error())
	}

	cache := &Cache{
		client: client,
		ctx:    ctx,
		items:  make(map[string]cacheItem),
	}

	// Автоматическая очистка просроченных ключей
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			cache.DeleteExpired()
		}
	}()

	return cache
}

func (c *Cache) Set(key string, value interface{}, expiration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var exp time.Time
	if expiration > 0 {
		exp = time.Now().Add(expiration)
	}

	c.items[key] = cacheItem{
		value:      value,
		expiration: exp,
	}

	log.Printf("Cache SET: key=%s, expiration=%v", key, expiration)

	jsonData, err := json.Marshal(value)
	if err != nil {
		log.Printf("Cache serialization error: %s", err.Error())
		return
	}

	err = c.client.Set(c.ctx, key, jsonData, expiration).Err()
	if err != nil {
		log.Printf("Cache set error: %s", err.Error())
	}
}

func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[key]
	if !found {
		log.Printf("Cache MISS: key=%s", key)
		return nil, false
	}

	if !item.expiration.IsZero() && item.expiration.Before(time.Now()) {
		log.Printf("Cache EXPIRED: key=%s", key)
		delete(c.items, key)
		return nil, false
	}

	log.Printf("Cache HIT: key=%s", key)

	jsonData, err := c.client.Get(c.ctx, key).Bytes()
	if err != nil {
		if err != redis.Nil {
			log.Printf("Cache get error: %s", err.Error())
		}
		return nil, false
	}

	var value interface{}
	err = json.Unmarshal(jsonData, &value)
	if err != nil {
		log.Printf("Cache deserialization error: %s", err.Error())
		return nil, false
	}

	return value, true
}

func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, found := c.items[key]; found {
		delete(c.items, key)
		log.Printf("Cache DELETE: key=%s", key)
	} else {
		log.Printf("Cache DELETE attempted but key not found: key=%s", key)
	}

	err := c.client.Del(c.ctx, key).Err()
	if err != nil {
		log.Printf("Cache delete error: %s", err.Error())
	}
}

func (c *Cache) DeleteExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	var expiredCount int

	for key, item := range c.items {
		if !item.expiration.IsZero() && item.expiration.Before(now) {
			delete(c.items, key)
			expiredCount++
		}
	}

	if expiredCount > 0 {
		log.Printf("Cache auto-cleanup: removed %d expired items", expiredCount)
	}
}

func (c *Cache) GetStats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := make(map[string]interface{})
	keys := make([]string, 0, len(c.items))

	for key := range c.items {
		keys = append(keys, key)
	}

	redisStats := make(map[string]interface{})
	if cmd := c.client.Info(c.ctx); cmd.Err() == nil {
		redisInfo := cmd.Val()
		redisStats["info"] = redisInfo
	}

	if cmd := c.client.DBSize(c.ctx); cmd.Err() == nil {
		redisStats["keys_count"] = cmd.Val()
	}

	stats["memory_keys"] = keys
	stats["memory_keys_count"] = len(c.items)
	stats["redis"] = redisStats

	return stats
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
