package redis

import (
    "context"
    "encoding/json"
    "errors"
    "time"

    "github.com/redis/go-redis/v9"
)

type IdempotencyStore struct {
    client *redis.Client
}

func NewIdempotencyStore(c *redis.Client) *IdempotencyStore {
    return &IdempotencyStore{client: c}
}

func (s *IdempotencyStore) Check(ctx context.Context, key string) ([]byte, bool, error) {
    val, err := s.client.Get(ctx, "idem:"+key).Bytes()
    if errors.Is(err, redis.Nil) {
        return nil, false, nil
    }
    if err != nil {
        return nil, false, err
    }
    return val, true, nil
}

func (s *IdempotencyStore) Store(ctx context.Context, key string, resp interface{}, ttl time.Duration) error {
    b, err := json.Marshal(resp)
    if err != nil {
        return err
    }
    return s.client.Set(ctx, "idem:"+key, b, ttl).Err()
}
