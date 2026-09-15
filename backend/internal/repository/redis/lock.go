package redis

import (
    "context"
    "fmt"
    "time"

    "github.com/google/uuid"
    "github.com/redis/go-redis/v9"
)

type Locker struct {
    client *redis.Client
}

func NewLocker(client *redis.Client) *Locker { return &Locker{client: client} }

func (l *Locker) Acquire(ctx context.Context, key string, ttl time.Duration) (func(), error) {
    token := uuid.New().String()
    fullKey := fmt.Sprintf("lock:%s", key)
    ok, err := l.client.SetNX(ctx, fullKey, token, ttl).Result()
    if err != nil {
        return nil, err
    }
    if !ok {
        return nil, fmt.Errorf("resource is locked, try again")
    }
    unlock := func() {
        script := `
            if redis.call("GET", KEYS[1]) == ARGV[1] then
                return redis.call("DEL", KEYS[1])
            end
            return 0
        `
        l.client.Eval(context.Background(), script, []string{fullKey}, token)
    }
    return unlock, nil
}
