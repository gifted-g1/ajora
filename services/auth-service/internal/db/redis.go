
package db

import (
    "context"
    "log"
    "time"

    "github.com/redis/go-redis/v9"
)

func NewRedisClient(addr, password string, db int) *redis.Client {
    client := redis.NewClient(&redis.Options{
        Addr:         addr,
        Password:     password,
        DB:           db,
        PoolSize:     100,
        MinIdleConns: 10,
        DialTimeout:  5 * time.Second,
        ReadTimeout:  3 * time.Second,
        WriteTimeout: 3 * time.Second,
    })

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := client.Ping(ctx).Err(); err != nil {
        log.Printf("Warning: Redis connection failed: %v", err)
    } else {
        log.Println("Redis connected successfully")
    }

    return client
}

